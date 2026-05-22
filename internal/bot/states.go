package bot

// setState устанавливает состояние пользователя
func (b *BotApp) setState(userID int64, state string) {
	b.userStates[userID] = state
}

// clearState очищает состояние
func (b *BotApp) clearState(userID int64) {
	delete(b.userStates, userID)
}

// getState возвращает состояние
func (b *BotApp) getState(userID int64) string {
	return b.userStates[userID]
}
