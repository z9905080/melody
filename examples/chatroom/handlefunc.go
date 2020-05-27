package main

import (
	"encoding/json"
	"github.com/z9905080/melody"
	"log"
)

// CommonInterfaceDeserilize 二次反序列化，讓Data的資料可以正常地轉化
func CommonInterfaceDeserilize(data interface{}, dataType interface{}) error {

	marshalData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err2 := json.Unmarshal(marshalData, &dataType)
	if err2 != nil {
		return err2
	}

	return nil
}

// getServerFuncMap WS的OperationCode要註冊進來並對應到相對的Func
// 這裡有參照 C# Reflection的一個核心概念
// 使用函數變數延遲執行的時機
func getServerFuncMap() map[OperationCode]func(*melody.Melody, *melody.Session, MessageST) {
	return map[OperationCode]func(*melody.Melody, *melody.Session, MessageST){
		Login:        HandleLogin,
		JoinChannel:  HandleJoinChannel,
		LeaveChannel: HandleLeaveChannel,
		SendMessage:  HandleSendMessage,
	}
}

func HandleLogin(server *melody.Melody, session *melody.Session, msg MessageST) {

	var (
		request   LoginMsgST
		loginResp LoginRespST
	)
	if unMarshalErr := CommonInterfaceDeserilize(msg.Data, &request); unMarshalErr != nil {
		log.Println("unMarshalErr:", unMarshalErr)
	} else {
		session.Set("userID", request.UserID)
		// subscribe topic
		session.AddSub(publicChannel)
		loginResp.IsSuccess = true
	}

	resp := MessageST{
		OpCode: Login,
		Data:   loginResp,
	}

	if respBytes, err := json.Marshal(resp); err != nil {
		log.Println(err)
	} else {
		session.Write(respBytes)
	}
}

func HandleJoinChannel(server *melody.Melody, session *melody.Session, msg MessageST) {

	var (
		request         JoinChannelMsgST
		joinChannelResp JoinChannelRespST
	)
	if unMarshalErr := CommonInterfaceDeserilize(msg.Data, &request); unMarshalErr != nil {
		log.Println("unMarshalErr:", unMarshalErr)
	} else {
		// subscribe topic
		session.AddSub(request.ChannelName)
		joinChannelResp.IsSuccess = true
	}

	resp := MessageST{
		OpCode: msg.OpCode,
		Data:   joinChannelResp,
	}

	if respBytes, err := json.Marshal(resp); err != nil {
		log.Println(err)
	} else {
		session.Write(respBytes)
	}
}

func HandleLeaveChannel(server *melody.Melody, session *melody.Session, msg MessageST) {

	var (
		request          LeaveChannelMsgST
		leaveChannelResp LeaveChannelRespST
	)
	if unMarshalErr := CommonInterfaceDeserilize(msg.Data, &request); unMarshalErr != nil {
		log.Println("unMarshalErr:", unMarshalErr)
	} else {
		// subscribe topic
		session.UnSub(request.ChannelName)
		leaveChannelResp.IsSuccess = true
	}

	resp := MessageST{
		OpCode: msg.OpCode,
		Data:   leaveChannelResp,
	}

	if respBytes, err := json.Marshal(resp); err != nil {
		log.Println(err)
	} else {
		session.Write(respBytes)
	}
}

func HandleSendMessage(server *melody.Melody, session *melody.Session, msg MessageST) {

	var request SendMessageMsgST
	if unMarshalErr := CommonInterfaceDeserilize(msg.Data, &request); unMarshalErr != nil {
		log.Println("unMarshalErr:", unMarshalErr)
	} else {
		var userID int
		if item, isExist := session.Get("userID"); isExist {
			if data, isReflectOK := item.(int); isReflectOK {
				userID = data
			}
		}

		resp := MessageST{
			OpCode: ReceiveMessage,
			Data: ReceiveMessageMsgST{
				UserID:      userID,              // 使用者ID
				ChannelName: request.ChannelName, // 告知使用者訊息來自哪個頻道
				Msg:         request.Msg,         // 訊息內容
			},
		}

		if respBytes, err := json.Marshal(resp); err != nil {
			log.Println(err)
		} else {
			server.PubMsg(respBytes, true, request.ChannelName)
		}
	}
}
