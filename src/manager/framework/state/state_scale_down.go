package state

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

type StateScaleDown struct {
	name string
	app  *App

	currentSlot      *Slot
	currentSlotIndex int
	targetSlotIndex  int
	lock             sync.Mutex
}

func NewStateScaleDown(app *App) *StateScaleDown {
	return &StateScaleDown{
		app:  app,
		name: APP_STATE_SCALE_DOWN,
	}
}

func (scaleDown *StateScaleDown) OnEnter() {
	logrus.Debug("state scaleDown OnEnter")

	scaleDown.app.EmitAppEvent(scaleDown.name)

	scaleDown.currentSlotIndex = len(scaleDown.app.GetSlots()) - 1
	scaleDown.targetSlotIndex = int(scaleDown.app.CurrentVersion.Instances)

	scaleDown.currentSlot = NewSlot(scaleDown.app, scaleDown.app.CurrentVersion, scaleDown.currentSlotIndex)

	scaleDown.currentSlot, _ = scaleDown.app.GetSlot(scaleDown.currentSlotIndex)
	scaleDown.currentSlot.KillTask()
}

func (scaleDown *StateScaleDown) OnExit() {
	logrus.Debug("state scaleDown OnExit")
}

func (scaleDown *StateScaleDown) Step() {
	logrus.Debug("state scaleDown step")

	if scaleDown.SlotSafeToRemoveFromApp(scaleDown.currentSlot) && scaleDown.currentSlotIndex == scaleDown.targetSlotIndex {
		scaleDown.app.RemoveSlot(scaleDown.currentSlotIndex)
		scaleDown.app.TransitTo(APP_STATE_NORMAL)
	} else if scaleDown.SlotSafeToRemoveFromApp(scaleDown.currentSlot) && (scaleDown.currentSlotIndex > scaleDown.targetSlotIndex) {
		scaleDown.lock.Lock()

		scaleDown.app.RemoveSlot(scaleDown.currentSlotIndex)
		scaleDown.currentSlotIndex -= 1
		scaleDown.currentSlot, _ = scaleDown.app.GetSlot(scaleDown.currentSlotIndex)
		scaleDown.currentSlot.KillTask()

		scaleDown.lock.Unlock()
	} else {
		logrus.Info("state scaleDown step, do nothing")
	}
}

func (scaleDown *StateScaleDown) SlotSafeToRemoveFromApp(slot *Slot) bool {
	return slot.StateIs(SLOT_STATE_REAP) || slot.Abnormal()
}

func (scaleDown *StateScaleDown) Name() string {
	return scaleDown.name
}

// state machine can transit to any state if current state is scaleDown
func (scaleDown *StateScaleDown) CanTransitTo(targetState string) bool {
	logrus.Debugf("state scaleDown CanTransitTo %s", targetState)

	return true
}
