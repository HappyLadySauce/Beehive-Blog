package ws

import "sync"

// hub 限制同一用户并发 WebSocket 连接数，避免误开多 Tab 耗尽资源。
type hub struct {
	mu         sync.Mutex
	perUser    map[int64]int
	maxPerUser int
}

func newHub(maxPerUser int) *hub {
	if maxPerUser <= 0 {
		maxPerUser = 5
	}
	return &hub{
		perUser:    make(map[int64]int),
		maxPerUser: maxPerUser,
	}
}

func (h *hub) tryAdd(userID int64) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.perUser[userID] >= h.maxPerUser {
		return false
	}
	h.perUser[userID]++
	return true
}

func (h *hub) remove(userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	n := h.perUser[userID] - 1
	if n <= 0 {
		delete(h.perUser, userID)
		return
	}
	h.perUser[userID] = n
}
