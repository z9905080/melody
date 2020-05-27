package main

type OperationCode int

const (
	Default        OperationCode = iota // 0為預設值
	Login                               // 登入
	JoinChannel                         // 加入頻道
	LeaveChannel                        // 離開頻道
	SendMessage                         // 發送訊息
	ReceiveMessage                      // 接收訊息
)

// MessageST message struct to send/receive client
type MessageST struct {
	OpCode OperationCode `json:"op_code"`
	Data   interface{}   `json:"data"`
}

type LoginMsgST struct {
	UserID int `json:"user_id"`
}

type LoginRespST struct {
	IsSuccess bool `json:"is_success"`
}

type JoinChannelMsgST struct {
	ChannelName string `json:"channel_name"`
}

type JoinChannelRespST struct {
	IsSuccess bool `json:"is_success"`
}

type LeaveChannelMsgST struct {
	ChannelName string `json:"channel_name"`
}

type LeaveChannelRespST struct {
	IsSuccess bool `json:"is_success"`
}

type SendMessageMsgST struct {
	ChannelName string `json:"channel_name"`
	Msg         string `json:"msg"`
}

type ReceiveMessageMsgST struct {
	UserID      int    `json:"user_id"`
	ChannelName string `json:"channel_name"`
	Msg         string `json:"msg"`
}
