package launch

type Status int

const (
	Starting Status = 0
	Started  Status = 1
	Stopping Status = 2
	Stopped  Status = 3
	Crashed  Status = 4
	Killed   Status = 5
)

func (s Status) String() (string) {
	switch s {
	case Starting:
		return "Starting"
	case Started:
		return "Started"
	case Stopping:
		return "Stopping"
	case Stopped:
		return "Stopped"
	case Crashed:
		return "Crashed"
	case Killed:
		return "Killed"
	default:
		return "Unknown"
	}
}
