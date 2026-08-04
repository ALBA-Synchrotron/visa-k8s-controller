package v1

type InstanceState string
type DesiredInstanceState string

const (
	StateActive    InstanceState = "ACTIVE"
	StateBuilding  InstanceState = "BUILDING"
	StateDeleting  InstanceState = "DELETING"
	StateError     InstanceState = "ERROR"
	StateStarting  InstanceState = "STARTING"
	StateStopped   InstanceState = "STOPPED"
	StateStopping  InstanceState = "STOPPING"
	StateRebooting InstanceState = "REBOOTING"
	StateUnknown   InstanceState = "UNKNOWN"
)

const (
	DesiredStateActive  = DesiredInstanceState(StateActive)
	DesiredStateStopped = DesiredInstanceState(StateStopped)
)
