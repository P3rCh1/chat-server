package models

import "time"

type WSResponse struct {
	Type string `json:"type"`
}

type WSRequest struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	NewRoomID int64  `json:"roomID"`
}

type Message struct {
	WSResponse
	ID        int64     `json:"id"`
	RoomID    int64     `json:"roomID"`
	UID       int64     `json:"uid"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

type WSError struct {
	WSResponse
	Error string `json:"error"`
}

type SentResponse struct {
	WSResponse
	MessageID int64     `json:"messageID"`
	Timestamp time.Time `json:"timestamp"`
}

func NewWSError(msg string) *WSError {
	return &WSError{
		WSResponse: WSResponse{Type: "error"},
		Error:      msg,
	}
}

func NewSentResponse(id int64, ts time.Time) *SentResponse {
	return &SentResponse{
		WSResponse: WSResponse{Type: "sent"},
		MessageID:  id,
		Timestamp:  ts,
	}
}

func NewEnterResponse() *WSResponse {
	return &WSResponse{Type: "enter"}
}
