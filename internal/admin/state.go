package admin

// AdminState хранит состояние админ-панели
type AdminState struct {
	Action   string
	Step     int
	EntityID uint
	TempData map[string]interface{}
}

// AdminFSM управляет состояниями всех админов
type AdminFSM struct {
	states map[int64]*AdminState
}

// Конструктор FSM
func NewAdminFSM() *AdminFSM {
	return &AdminFSM{
		states: make(map[int64]*AdminState),
	}
}

// Получить состояние
func (fsm *AdminFSM) GetState(userID int64) (*AdminState, bool) {
	state, exists := fsm.states[userID]
	return state, exists
}

// Установить состояние
func (fsm *AdminFSM) SetState(userID int64, state *AdminState) {
	fsm.states[userID] = state
}

// Удалить состояние
func (fsm *AdminFSM) DeleteState(userID int64) {
	delete(fsm.states, userID)
}
