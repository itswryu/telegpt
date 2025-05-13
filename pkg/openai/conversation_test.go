package openai

import (
	"testing"

	"github.com/swryu/telegpt/pkg/config"
)

func TestConversationHistory(t *testing.T) {
	// 테스트를 위한 클라이언트 생성
	cfg := &config.Config{
		OpenAI: config.OpenAIConfig{
			APIKey: "test-key",
			Model:  "gpt-4.1-nano",
		},
	}
	client := NewClient(cfg)

	// 테스트용 사용자 ID
	userID := int64(12345)

	// 초기 대화 상태 확인
	// 대화가 존재하지 않으면 GetConversation에서 새로운 대화를 생성하므로
	// 대화 기록이 비어있는지 확인
	conv := client.convManager.GetConversation(userID)
	if len(conv.Messages) > 0 {
		t.Errorf("대화가 비어있어야 함")
	}

	// 첫 번째 메시지 처리 시뮬레이션 (실제 API 호출 없이)
	message1 := "안녕하세요"
	client.addMessageToHistory(userID, "user", message1)

	// 대화 기록에 메시지가 추가되었는지 확인
	conv = client.convManager.GetConversation(userID)

	if len(conv.Messages) != 1 {
		t.Errorf("대화 기록에 메시지가 1개 있어야 함, 현재: %d", len(conv.Messages))
	}

	if conv.Messages[0].Role != "user" || conv.Messages[0].Content != message1 {
		t.Errorf("잘못된 메시지 저장됨: %+v", conv.Messages[0])
	}

	// 봇 응답 추가
	botResponse := "안녕하세요! 무엇을 도와드릴까요?"
	client.addMessageToHistory(userID, "assistant", botResponse)

	// 두 번째 사용자 메시지 추가
	message2 := "내 이름을 기억해줘. 내 이름은 철수야."
	client.addMessageToHistory(userID, "user", message2)

	// 두 번째 봇 응답 추가
	botResponse2 := "네, 철수님! 기억하겠습니다."
	client.addMessageToHistory(userID, "assistant", botResponse2)

	// 대화 기록 확인
	conv = client.convManager.GetConversation(userID)
	messageCount := len(conv.Messages)

	if messageCount != 4 {
		t.Errorf("대화 기록에 메시지가 4개 있어야 함, 현재: %d", messageCount)
	}

	// 세 번째 사용자 메시지 추가
	message3 := "내 이름이 뭐지?"
	client.addMessageToHistory(userID, "user", message3)

	// 전체 대화 기록 가져오기
	conv = client.convManager.GetConversation(userID)
	messages := make([]Message, len(conv.Messages))
	copy(messages, conv.Messages)

	// 대화 기록에 모든 메시지가 순서대로 저장되어 있는지 확인
	expectedMessages := []struct {
		role    string
		content string
	}{
		{"user", message1},
		{"assistant", botResponse},
		{"user", message2},
		{"assistant", botResponse2},
		{"user", message3},
	}

	if len(messages) != len(expectedMessages) {
		t.Errorf("예상 메시지 수: %d, 실제: %d", len(expectedMessages), len(messages))
	}

	for i, expected := range expectedMessages {
		if i >= len(messages) {
			break
		}
		if messages[i].Role != expected.role || messages[i].Content != expected.content {
			t.Errorf("메시지 #%d: 예상 - {%s: %s}, 실제 - {%s: %s}",
				i, expected.role, expected.content, messages[i].Role, messages[i].Content)
		}
	}

	// 대화 초기화 테스트
	client.ResetConversation(userID)

	conv = client.convManager.GetConversation(userID)

	if len(conv.Messages) != 0 {
		t.Errorf("대화 초기화 후 메시지가 없어야 함, 현재: %d", len(conv.Messages))
	}

	// 메시지 최대 개수 제한 테스트
	// 최대 개수보다 많은 메시지를 추가하고 가장 오래된 메시지가 제거되는지 테스트
	for i := 0; i < maxHistory+5; i++ {
		client.addMessageToHistory(userID, "user", "테스트 메시지 "+string(rune('A'+i)))
	}

	conv = client.convManager.GetConversation(userID)
	messages = make([]Message, len(conv.Messages))
	copy(messages, conv.Messages)

	if len(messages) > maxHistory {
		t.Errorf("메시지 수가 최대값(%d)을 초과함: %d", maxHistory, len(messages))
	}

	// 첫 5개 메시지가 제거되었는지 확인
	firstMessage := messages[0].Content
	if firstMessage != "테스트 메시지 "+string(rune('A'+5)) {
		t.Errorf("오래된 메시지가 제거되지 않음: %s", firstMessage)
	}
}

// 이제 addMessageToHistory 메서드는 openai.go 파일에 구현되어 있음
