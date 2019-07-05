package melody

// register
// topics    Key: topic  , Value: 有訂閱此Topic的ChannelMap
// revTopics Key: Channel, Value: 訂閱了哪些Topic
type register struct {
	topics    map[string]map[chan interface{}]bool
	revTopics map[chan interface{}]map[string]bool
}

func (reg *register) add(topic string, ch chan interface{}) {
	if reg.topics[topic] == nil {
		reg.topics[topic] = make(map[chan interface{}]bool)
	}
	reg.topics[topic][ch] = true

	if reg.revTopics[ch] == nil {
		reg.revTopics[ch] = make(map[string]bool)
	}
	reg.revTopics[ch][topic] = true
}

func (reg *register) send(topic string, msg interface{}) {
	for ch := range reg.topics[topic] {
		ch <- msg
	}
}

func (reg *register) sendAsync(topic string, msg interface{}) {
	for ch := range reg.topics[topic] {
		select {
		case ch <- msg:
		default:
		}

	}
}

func (reg *register) removeTopic(topic string) {
	for ch := range reg.topics[topic] {
		reg.remove(topic, ch)
	}
}

func (reg *register) removeChannel(ch chan interface{}) {
	for topic := range reg.revTopics[ch] {
		reg.remove(topic, ch)
	}
}

func (reg *register) remove(topic string, ch chan interface{}) {
	if _, ok := reg.topics[topic]; !ok {
		return
	}

	if _, ok := reg.topics[topic][ch]; !ok {
		return
	}

	delete(reg.topics[topic], ch)
	delete(reg.revTopics[ch], topic)

	if len(reg.topics[topic]) == 0 {
		delete(reg.topics, topic)
	}

	if len(reg.revTopics[ch]) == 0 {
		close(ch)
		delete(reg.revTopics, ch)
	}
}
