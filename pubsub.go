package melody

type operation int

const (
	// Subscribe 訂閱
	Subscribe operation = iota
	// Publish 發布訊息
	Publish
	// AsyncPublish 非同步發布訊息
	AsyncPublish
	// Unsubscribe 取消訂閱
	Unsubscribe
	// UnSubscribeAll 全部取消訂閱
	UnSubscribeAll
	// CloseTopic 關閉此標題
	CloseTopic
	// ShutDown 此訂閱服務關機
	ShutDown
)

// pubSubPattern 集合topic,capacity是容量
type pubSub struct {
	commandChan chan cmd // 接收指令的channel
	bufferSize  int
}

type cmd struct {
	opCode operation      // 指令
	topics []string       // 訂閱的主題
	ch     chan *envelope // 使用的channel
	msg    *envelope      // 訊息內文
}

// pubSubNew 創建一個訂閱者模式(pattern)
func pubSubNew(bufferSize int) *pubSub {
	ps := &pubSub{make(chan cmd), bufferSize}
	go ps.start()
	return ps
}

// Sub 創建一個新的訂閱頻道, 並將channel回傳
func (ps *pubSub) Sub(topics ...string) chan *envelope {
	return ps.sub(Subscribe, topics...)
}

func (ps *pubSub) sub(op operation, topics ...string) chan *envelope {
	ch := make(chan *envelope, ps.bufferSize)
	ps.commandChan <- cmd{opCode: op, topics: topics, ch: ch}
	return ch
}

// AddSub 將要訂閱的Topic加到現有的channel
func (ps *pubSub) AddSub(ch chan *envelope, topics ...string) {
	ps.commandChan <- cmd{opCode: Subscribe, topics: topics, ch: ch}
}

// Pub 發布訊息
func (ps *pubSub) Pub(msg *envelope, topics ...string) {
	ps.commandChan <- cmd{opCode: Publish, topics: topics, msg: msg}
}

// AsyncPub 非同步的發布訊息，內部機制（for select default）
func (ps *pubSub) AsyncPub(msg *envelope, topics ...string) {
	ps.commandChan <- cmd{opCode: AsyncPublish, topics: topics, msg: msg}
}

// Unsub 取消訂閱
func (ps *pubSub) Unsub(ch chan *envelope, topics ...string) {
	// 如果不寫topic，視為將全部topic都取消訂閱
	if len(topics) == 0 {
		ps.commandChan <- cmd{opCode: UnSubscribeAll, ch: ch}
		return
	}

	ps.commandChan <- cmd{opCode: Unsubscribe, topics: topics, ch: ch}
}

// Close 關閉Topic, 相關有訂閱的channel都會被取消
func (ps *pubSub) Close(topics ...string) {
	ps.commandChan <- cmd{opCode: CloseTopic, topics: topics}
}

// Shutdown 關閉所有有訂閱的Channel
func (ps *pubSub) Shutdown() {
	ps.commandChan <- cmd{opCode: ShutDown}
}

func (ps *pubSub) start() {

	// 初始化暫存在記憶體的資料(topicsMap & revertTopicsOfChannelMap)
	reg := register{
		topics:    make(map[string]map[chan *envelope]bool),
		revTopics: make(map[chan *envelope]map[string]bool),
	}

loop:
	for cmd := range ps.commandChan {
		if cmd.topics == nil {
			switch cmd.opCode {
			case UnSubscribeAll:
				reg.removeChannel(cmd.ch)

			case ShutDown:
				break loop
			}

			continue loop
		}

		for _, topic := range cmd.topics {
			switch cmd.opCode {
			case Subscribe:
				reg.add(topic, cmd.ch)

			case Publish:
				reg.send(topic, cmd.msg)

			case AsyncPublish:
				reg.sendAsync(topic, cmd.msg)

			case Unsubscribe:
				reg.remove(topic, cmd.ch)

			case CloseTopic:
				reg.removeTopic(topic)
			}
		}
	}

	// 當跳出迴圈要結束時，將所有未關閉的topic channel進行移除
	for topic, chans := range reg.topics {
		for ch := range chans {
			reg.remove(topic, ch)
		}
	}
}
