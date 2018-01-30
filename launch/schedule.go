package launch

import "time"

func doSchedule(m *Manager, cancelled *bool) {
	if m.config.Schedule.Restart < 0 {
		return
	}

	var start = time.Now()
	var previous = start
	var notifyPeriod = time.Duration(5) * time.Minute

	for {
		if *cancelled {
			return
		}

		// process is running normally
		if m.getStatus() == started {
			elapsed := time.Since(start)
			remaining := m.getRestartWait() - elapsed

			// no time remaining so restart
			if remaining <= 0 {
				m.Restart()
				return
			}

			// notify every minute for the last 5 minutes (before restarting)
			if remaining < notifyPeriod {
				if time.Since(previous) > time.Minute {
					m.Say("Scheduled restart in %.0f minutes", remaining.Minutes())
					previous = time.Now()
				}
			}
		}

		time.Sleep(time.Second)
	}
}