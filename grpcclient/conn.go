package conn

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	aiyypb "aisi/module/aiyypb"
	msgsrv "aisi/module/msg"
	streampb "aisi/module/stream"
	"aisi/module/trans"
	"aisi/utils"

	log "aitools/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	aiyyaddr string
	rtaddr   string
	connmap  sync.Map
	timeout  int64
)

const (
	// 连接正常
	VALID = iota
	// 连接超时
	TIME_OUT
)

type (
	// ConnInfo 表示streamplugin和aiyysrv的连接实例
	ConnInfo struct {
		rtmpaddr      string
		topic         string
		flag          int32
		msgid         int
		rtmpconn      *grpc.ClientConn
		rtmpclient    streampb.StreamClient
		rtmpstream    streampb.Stream_CreateStreamClient
		aiyyconn      *grpc.ClientConn
		aiyyclient    aiyypb.AiyyClient
		aiyystream    aiyypb.Aiyy_PostClient
		aiyyclosechan chan bool
		once          sync.Once
	}
	// RecoRes 识别结果，并附加msgid
	RecoRes struct {
		*aiyypb.RspAiyy_Result
		Msgid int
	}
)

// Init 初始化模块变量
func Init(aiyyAddr string, rtAddr string, _timeout int64) {
	aiyyaddr = aiyyAddr
	rtaddr = rtAddr
	timeout = _timeout
	go CheckTimeout()
}

// CheckTimeout 检查模块中保存的conninfo实例是否有超时
func CheckTimeout() {
	checkinterval := time.Duration(timeout) * time.Second
	for {
		time.Sleep(checkinterval)
		needStop := make([]*ConnInfo, 0)
		connmap.Range(func(key, value interface{}) bool {
			ci, ok := value.(*ConnInfo)
			if !ok {
				log.Errorf("CheckTimeout:expect *ConnInfo, got %T", value)
				return true
			}
			if atomic.LoadInt32(&ci.flag) == TIME_OUT {
				log.Infof("CheckTimeout:found timeout connection,prepare to stop:%s", ci.rtmpaddr)
				//ci.stop()
				needStop = append(needStop, ci)
			} else {
				atomic.StoreInt32(&ci.flag, TIME_OUT)
			}
			return true
		})

		for _, ci := range needStop {
			ci.stop()
		}
	}
}

//send audio data
func (c *ConnInfo) send() {
	log.Info("rtmp send goroutine started:", c.rtmpaddr)
	p := new(aiyypb.ReqAiyy_Package)
	var err error
	var msg *streampb.Package
	for {
		msg, err = c.rtmpstream.Recv()
		if err != nil {
			log.Errorf("fail to receive data from rtmp parse server:%v", err)
			c.stop()
			break
		}
		atomic.StoreInt32(&c.flag, VALID)

		//fmt.Println(msg.Pkgid)
		p.Package = &aiyypb.Package{Data: msg.Data, Pkgid: uint64(msg.Pkgid)}
		err = c.aiyystream.Send(&aiyypb.ReqAiyy{Body: p})
		if err != nil {
			log.Errorf("fail to send data to aiyyserver:%v", err)
			c.stop()
			break
		}
	}
	err = c.aiyystream.CloseSend()
	if err != nil {
		log.Errorf("fail to CloseSend aiyy stream:%s, error:%v", c.rtmpaddr, err)
	}
	log.Infof("rtmp send goroutine stopped:%v", c.rtmpaddr)
	close(c.aiyyclosechan)
}

//receive recognize result
func (c *ConnInfo) receive() {
	log.Infof("rtmp receive goroutine started:%v", c.rtmpaddr)
	var err error
	var m *aiyypb.RspAiyy
	for {
		m, err = c.aiyystream.Recv()
		if err != nil {
			log.Errorf("fail to receive data from aiyysrv:%v", err)
			c.stop()
			break
		}
		switch body := m.Body.(type) {
		case *aiyypb.RspAiyy_Result:
			msgsrv.Push(c.topic, msgsrv.Message{Type: 0, Data: RecoRes{body, c.msgid}})
			log.Infof("%d-%d:%s %s\n", body.Result.StartTime, body.Result.EndTime, body.Result.Text, body.Result.Sinfo)
			if body.Result.StatusCode == 0 {
				trans.Translate(trans.Trans{MeetingId: c.topic, SentenceId: c.msgid, Text: body.Result.Text})
			}
		case *aiyypb.RspAiyy_Msg:
			log.Infof("grpc error: %d:%s\n", body.Msg.Code, body.Msg.Msg)
			c.stop()
			break
		}
		c.msgid++
	}
	<-c.aiyyclosechan
	err = c.aiyyconn.Close()
	if err != nil {
		log.Errorf("fail to Close aiyy grpc conn:%s, error:%v", c.rtmpaddr, err)
	}
	msgsrv.Push(c.topic, msgsrv.Message{Type: 9, Data: RecoRes{}})
	log.Infof("rtmp receive goroutine stopped:%v", c.rtmpaddr)
}

// stop 方法会停止并断开语音识别相关进程
// 该方法会在以下地方发生调用：
// 1. 主动调用stop方法，比如超时或者收到关闭请求
// 2. 解析rmtp的grpc连接recv收到错误
// 3. aiyy的grpc连接send的时候出现错误
// 4. aiyy的grpc连接recv的时候出现错误
// 5. 收到aiyy的msg信息
func (c *ConnInfo) stop() {
	c.once.Do(c.realstop)
}

func (c *ConnInfo) realstop() {
	connmap.Delete(c.rtmpaddr)
	//close allclient
	var err error
	err = c.rtmpconn.Close()
	if err != nil {
		log.Errorf("fail to Close rtmp grpc conn:%s, error:%v", c.rtmpaddr, err)
	}
}

// NewStream 新建ConnInfo实例，如果已经存在会直接返回
func NewStream(rtmpaddr string, custopic string) (string, error) {
	ciItf, ok := connmap.Load(rtmpaddr)
	if ok {
		ci, ok := ciItf.(*ConnInfo)
		if !ok {
			log.Errorf("expect *ConnInfo,got %T", ciItf)
			return "", errors.New("convert error")
		}
		return ci.topic, nil
	}

	ci := &ConnInfo{}
	ci.msgid = 0
	ci.aiyyclosechan = make(chan bool)
	ci.rtmpaddr = rtmpaddr
	ci.flag = VALID
	if custopic == "" {
		ci.topic = utils.GetTopic(rtmpaddr)
	} else {
		ci.topic = custopic
	}
	//connect to rtmp parse server
	conn, err := grpc.Dial(rtaddr, grpc.WithInsecure())
	if err != nil {
		return ci.topic, err
	}
	ci.rtmpconn = conn

	client := streampb.NewStreamClient(conn)
	ci.rtmpclient = client
	info := new(streampb.StreamInfo)
	info.Addr = rtmpaddr
	info.Opt = ""
	info.Scheme = streampb.StreamInfo_RTMP
	stream, err := client.CreateStream(context.Background(), info)
	if err != nil {
		return ci.topic, err
	}
	ci.rtmpstream = stream

	//connect to aiyydb
	yyconn, err := grpc.Dial(aiyyaddr, grpc.WithInsecure())
	if err != nil {
		return ci.topic, err
	}
	ci.aiyyconn = yyconn
	yyclient := aiyypb.NewAiyyClient(yyconn)
	ci.aiyyclient = yyclient
	ctx := metadata.AppendToOutgoingContext(context.Background(), "sid", "fdasg")
	yystream, err := yyclient.Post(ctx)
	if err != nil {
		return ci.topic, err
	}
	ci.aiyystream = yystream
	_, loaded := connmap.LoadOrStore(rtmpaddr, ci)
	if !loaded {
		go ci.send()
		go ci.receive()
	}
	return ci.topic, nil
}

// StopStream 停止某个ConnInfo连接实例
func StopStream(rtmpaddr string) {
	ciItf, ok := connmap.Load(rtmpaddr)
	if !ok {
		return
	}
	ci, ok := ciItf.(*ConnInfo)
	if !ok {
		log.Errorf("StopStream: expect *ConnInfo,got %T", ciItf)
		return
	}
	ci.stop()
	return
}

// StatStream 检查某个连接实例是否存在
func StatStream(rtmpaddr string) string {
	_, ok := connmap.Load(rtmpaddr)
	if !ok {
		return "not found"
	}
	return "ok"
}
