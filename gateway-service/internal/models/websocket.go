package models

import "time"

type WSResponse struct {
	Type string `json:"Type"`
}

type WSRequest struct {
	Type      string `json:"Type"`
	Text      string `json:"Text"`
	NewRoomID int64  `json:"RoomID"`
}

type Message struct {
	WSResponse
	ID        int64     `json:"ID"`
	RoomID    int64     `json:"RoomID"`
	UID       int64     `json:"UID"`
	Text      string    `json:"Text"`
	Timestamp time.Time `json:"Timestamp"`
}

type WSError struct {
	WSResponse
	Error string `json:"Error"`
}

type SentResponse struct {
	WSResponse
	MessageID int64     `json:"MessageID"`
	Timestamp time.Time `json:"Timestamp"`
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
