package status

type InstanceState string
type DesiredInstanceState string

const (
	StateActive      InstanceState = "ACTIVE"
	StateBuilding    InstanceState = "BUILDING"
	StateDeleting    InstanceState = "DELETING"
	StateError       InstanceState = "ERROR"
	StateStarting    InstanceState = "STARTING"
	StateStopped     InstanceState = "STOPPED"
	StateStopping    InstanceState = "STOPPING"
	StateRebooting   InstanceState = "REBOOTING"
	StateUnavailable InstanceState = "UNAVAILABLE"
	StateUnknown     InstanceState = "UNKNOWN"
)

const (
	DesiredStateActive  DesiredInstanceState = DesiredInstanceState(StateActive)
	DesiredStateStopped DesiredInstanceState = DesiredInstanceState(StateStopped)
)
