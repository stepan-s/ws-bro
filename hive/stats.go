package hive

import "sync"

type UsersStatsData struct {
	TotalConnectionsAccepted uint64
	CurrentConnections       uint32
	TotalUsersConnected      uint64
	CurrentUsersConnected    uint32
	MessagesReceived         uint64
	MessagesTransmitted      uint64
}

type UsersStats struct {
	inData  UsersStatsData
	out     chan UsersStatsData
	outData UsersStatsData
	lock    *sync.RWMutex
}

func NewUsersStats() *UsersStats {
	s := &UsersStats{
		out:  make(chan UsersStatsData, 10),
		lock: &sync.RWMutex{},
	}
	go func() {
		for {
			select {
			case data := <-s.out:
				s.setData(data)
			}
		}
	}()
	return s
}

func (s *UsersStats) sync() {
	if len(s.out) < cap(s.out) {
		s.out <- s.inData
	}
}

func (s *UsersStats) setData(data UsersStatsData) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.outData = data
}

func (s *UsersStats) Connected() {
	s.inData.TotalUsersConnected += 1
	s.inData.CurrentUsersConnected += 1
	s.inData.TotalConnectionsAccepted += 1
	s.inData.CurrentConnections += 1
	s.sync()
}

func (s *UsersStats) ConnectionAdded() {
	s.inData.TotalConnectionsAccepted += 1
	s.inData.CurrentConnections += 1
	s.sync()
}

func (s *UsersStats) ConnectionRemoved() {
	s.inData.CurrentConnections -= 1
	s.sync()
}

func (s *UsersStats) Disconnected() {
	s.inData.CurrentConnections -= 1
	s.inData.CurrentUsersConnected -= 1
	s.sync()
}

func (s *UsersStats) Received() {
	s.inData.MessagesReceived += 1
	s.sync()
}

func (s *UsersStats) Transmitted() {
	s.inData.MessagesTransmitted += 1
	s.sync()
}

func (s *UsersStats) GetData() UsersStatsData {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.inData
}

type AppsStatsData struct {
	TotalConnectionsAccepted uint64
	TotalReconnects          uint64
	TotalDisconnects         uint64
	CurrentConnections       uint32
	MessagesReceived         uint64
	MessagesTransmitted      uint64
}

type AppsStats struct {
	inData  AppsStatsData
	out     chan AppsStatsData
	outData AppsStatsData
	lock    *sync.RWMutex
}

func NewAppsStats() *AppsStats {
	s := &AppsStats{
		out:  make(chan AppsStatsData, 10),
		lock: &sync.RWMutex{},
	}
	go func() {
		for {
			select {
			case data := <-s.out:
				s.setData(data)
			}
		}
	}()
	return s
}

func (s *AppsStats) sync() {
	if len(s.out) < cap(s.out) {
		s.out <- s.inData
	}
}

func (s *AppsStats) setData(data AppsStatsData) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.outData = data
}

func (s *AppsStats) Connected() {
	s.inData.TotalConnectionsAccepted += 1
	s.inData.CurrentConnections += 1
	s.sync()
}

func (s *AppsStats) Disconnected() {
	s.inData.TotalDisconnects += 1
	s.inData.CurrentConnections -= 1
	s.sync()
}

func (s *AppsStats) Reconnected() {
	s.inData.TotalConnectionsAccepted += 1
	s.inData.TotalReconnects += 1
	s.sync()
}

func (s *AppsStats) Received() {
	s.inData.MessagesReceived += 1
	s.sync()
}

func (s *AppsStats) Transmitted() {
	s.inData.MessagesTransmitted += 1
	s.sync()
}

func (s *AppsStats) GetData() AppsStatsData {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.inData
}
