package p2p

import (
	"testing"
	"time"
)

func TestPeerScorer_InitialScore(t *testing.T) {
	ps := NewPeerScorer(nil)
	nodeID := randomNodeID()

	score := ps.GetScore(nodeID)
	if score != 0.5 {
		t.Fatalf("expected initial score 0.5, got %f", score)
	}
}

func TestPeerScorer_ValidMessageIncrease(t *testing.T) {
	ps := NewPeerScorer(nil)
	nodeID := randomNodeID()

	ps.RecordEvent(nodeID, PeerEventValidMessage)
	score := ps.GetScore(nodeID)
	if score <= 0.5 {
		t.Fatalf("expected score > 0.5 after valid message, got %f", score)
	}
}

func TestPeerScorer_InvalidMessageDecrease(t *testing.T) {
	ps := NewPeerScorer(nil)
	nodeID := randomNodeID()

	ps.RecordEvent(nodeID, PeerEventInvalidMessage)
	score := ps.GetScore(nodeID)
	if score >= 0.5 {
		t.Fatalf("expected score < 0.5 after invalid message, got %f", score)
	}

	peerScore := ps.GetPeerScore(nodeID)
	if peerScore.InvalidMessages != 1 {
		t.Fatalf("expected 1 invalid message counter, got %d", peerScore.InvalidMessages)
	}
}

func TestPeerScorer_ScoreClamping(t *testing.T) {
	ps := NewPeerScorer(nil)
	nodeID := randomNodeID()

	// Many valid messages should not exceed 1.0
	for i := 0; i < 1000; i++ {
		ps.RecordEvent(nodeID, PeerEventValidMessage)
	}
	score := ps.GetScore(nodeID)
	if score > 1.0 {
		t.Fatalf("score should not exceed 1.0, got %f", score)
	}

	// Many invalid messages should not go below 0.0
	for i := 0; i < 100; i++ {
		ps.RecordEvent(nodeID, PeerEventInvalidMessage)
	}
	score = ps.GetScore(nodeID)
	if score < 0.0 {
		t.Fatalf("score should not go below 0.0, got %f", score)
	}
}

func TestPeerScorer_BanThreshold(t *testing.T) {
	banList := NewBanList()
	ps := NewPeerScorer(nil)
	ps.SetBanList(banList)

	nodeID := randomNodeID()

	// Repeatedly trigger invalid messages to drop below ban threshold (0.1)
	// Start at 0.5, each -0.05 → need ~10 events to reach 0.0
	for i := 0; i < 10; i++ {
		ps.RecordEvent(nodeID, PeerEventInvalidMessage)
	}

	if !banList.IsBanned(nodeID) {
		score := ps.GetScore(nodeID)
		t.Fatalf("expected peer to be banned (score=%f, threshold=0.1)", score)
	}
}

func TestPeerScorer_DecayTowardNeutral(t *testing.T) {
	ps := NewPeerScorer(nil)
	nodeID := randomNodeID()

	// Drop score
	for i := 0; i < 5; i++ {
		ps.RecordEvent(nodeID, PeerEventInvalidMessage)
	}

	scoreBefore := ps.GetScore(nodeID)
	if scoreBefore >= 0.5 {
		t.Fatalf("expected score < 0.5, got %f", scoreBefore)
	}

	// Decay should bring it back toward 0.5
	ps.DecayScores()

	scoreAfter := ps.GetScore(nodeID)
	if scoreAfter <= scoreBefore {
		t.Fatalf("expected score to increase after decay, before=%f, after=%f", scoreBefore, scoreAfter)
	}
}

func TestPeerScorer_DecayAboveNeutral(t *testing.T) {
	ps := NewPeerScorer(nil)
	nodeID := randomNodeID()

	// Boost score
	for i := 0; i < 100; i++ {
		ps.RecordEvent(nodeID, PeerEventValidMessage)
	}

	scoreBefore := ps.GetScore(nodeID)
	if scoreBefore <= 0.5 {
		t.Fatalf("expected score > 0.5, got %f", scoreBefore)
	}

	// Decay should bring it back toward 0.5
	ps.DecayScores()

	scoreAfter := ps.GetScore(nodeID)
	if scoreAfter >= scoreBefore {
		t.Fatalf("expected score to decrease after decay, before=%f, after=%f", scoreBefore, scoreAfter)
	}
}

func TestPeerScorer_RemovePeer(t *testing.T) {
	ps := NewPeerScorer(nil)
	nodeID := randomNodeID()

	ps.RecordEvent(nodeID, PeerEventValidMessage)
	if ps.PeerCount() != 1 {
		t.Fatalf("expected 1 peer, got %d", ps.PeerCount())
	}

	ps.RemovePeer(nodeID)
	if ps.PeerCount() != 0 {
		t.Fatalf("expected 0 peers, got %d", ps.PeerCount())
	}

	// Should return default score
	score := ps.GetScore(nodeID)
	if score != 0.5 {
		t.Fatalf("expected default score after removal, got %f", score)
	}
}

func TestPeerScorer_AllEventTypes(t *testing.T) {
	ps := NewPeerScorer(nil)
	nodeID := randomNodeID()

	events := []PeerEvent{
		PeerEventValidMessage,
		PeerEventInvalidMessage,
		PeerEventFailedDeploy,
		PeerEventGossipSpam,
		PeerEventRateLimitHit,
		PeerEventMalformedPayload,
		PeerEventGoodUptime,
	}

	for _, event := range events {
		ps.RecordEvent(nodeID, event)
	}

	peerScore := ps.GetPeerScore(nodeID)
	if peerScore == nil {
		t.Fatal("expected non-nil peer score")
	}
}

func TestPeerScorer_NowFunc(t *testing.T) {
	now := time.Now()
	ps := NewPeerScorer(nil)
	ps.nowFunc = func() time.Time { return now }

	nodeID := randomNodeID()
	ps.RecordEvent(nodeID, PeerEventValidMessage)

	peerScore := ps.GetPeerScore(nodeID)
	if peerScore.LastUpdate != now {
		t.Fatalf("expected LastUpdate=%v, got %v", now, peerScore.LastUpdate)
	}
}
