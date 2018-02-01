package launch

import "time"

func scheduleRestart(m *Manager) {
	if m.config.Schedule.Restart < 0 {
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// wait until process has exited the starting phase
	for range ticker.C {
		if m.getStatus() != Starting {
			break
		}
	}

	start := time.Now()
	previous := start
	notifyPeriod := time.Duration(5) * time.Minute

	// commence schedule loop - repeat until the process is not in the running (started) phase
	for range ticker.C {
		if m.getStatus() != Started {
			return
		}

		// time since the process started
		elapsed := time.Since(start)

		// time remaining before restart should occur
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
}