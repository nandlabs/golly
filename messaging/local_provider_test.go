package messaging

import (
	"net/url"
	"testing"
)

func TestLocalProvider_Send(t *testing.T) {
	lms := Get()
	msg, err := lms.NewMessage("chan")
	input := "this is a test string"
	_, err = msg.SetBodyStr(input)
	if err != nil {
		t.Errorf("Error SetBodyStr:: %v", err)
	}
	uri, _ := url.Parse("chan://localhost:8080")
	got := lms.Send(uri, msg)
	if got != nil {
		t.Errorf("Error got :: %v", got)
	}

	uriErr, _ := url.Parse("http://localhost:8080")
	got = lms.Send(uriErr, msg)
	if got.Error() != "unsupported scheme http" {
		t.Errorf("Error got :: %v", got)
	}
}

func TestLocalProvider_SendBatch(t *testing.T) {
	lms := Get()
	msg1, err := lms.NewMessage("chan")
	input1 := "this is a test string 1"
	_, err = msg1.SetBodyStr(input1)
	if err != nil {
		t.Errorf("Error SetBodyStr:: %v", err)
	}
	msg2, err := lms.NewMessage("chan")
	input2 := "this is a test string 2"
	_, err = msg2.SetBodyStr(input2)
	if err != nil {
		t.Errorf("Error SetBodyStr:: %v", err)
	}
	uri, _ := url.Parse("chan://localhost:8080")
	msgs := []Message{msg1, msg2}
	got := lms.SendBatch(uri, msgs)
	if got != nil {
		t.Errorf("Error got :: %v", got)
	}
}

func TestLocalProvider_AddListener(t *testing.T) {
	lms := Get()
	uri, _ := url.Parse("chan://localhost2:8080")
	input1 := "this is a listener test"
	go func() {
		err := lms.AddListener(uri, func(output Message) {
			if output.ReadAsStr() != input1 {
				t.Errorf("Error AddListner")
			}
		})
		if err != nil {
			t.Errorf("Error AddListner")
		}
	}()

	msg1, err := lms.NewMessage("chan")
	_, err = msg1.SetBodyStr(input1)
	if err != nil {
		t.Errorf("Error SetBodyStr:: %v", err)
	}
	_ = lms.Send(uri, msg1)
}
