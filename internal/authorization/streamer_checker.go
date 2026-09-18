package authorization

type StreamerChecker struct {
	streamerTwitchUserID string
}

func NewStreamerChecker(streamerTwitchUserID string) StreamerChecker {
	return StreamerChecker{
		streamerTwitchUserID: streamerTwitchUserID,
	}
}

func (sc StreamerChecker) IsStreamer(twitchUserID string) bool {
	return sc.streamerTwitchUserID == twitchUserID
}
