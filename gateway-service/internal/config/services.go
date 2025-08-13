package config

import "time"

type Services struct {
	SessionAddr string           `yaml:"session_addr"`
	UserAddr    string           `yaml:"user_addr"`
	RoomsAddr   string           `yaml:"rooms_addr"`
	MessageAddr string           `yaml:"message_addr"`
	Timeouts    TimeoutsServices `yaml:"timeouts"`
}

type TimeoutsServices struct {
	Session time.Duration `yaml:"session"`
	User    time.Duration `yaml:"user"`
	Rooms   time.Duration `yaml:"rooms"`
	Message time.Duration `yaml:"message"`
}

func DefaultServices() Services {
	return Services{
		SessionAddr: "session:50051",
		UserAddr:    "user:50052",
		RoomsAddr:   "rooms:50053",
		MessageAddr: "message:50054",
		Timeouts: TimeoutsServices{
			Session: 3 * time.Second,
			User:    3 * time.Second,
			Rooms:   3 * time.Second,
			Message: 3 * time.Second,
		},
	}
}
