package pusher

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTriggerSuccessCase(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		fmt.Fprintf(res, "{}")
		assert.Equal(t, "POST", req.Method)

		expectedBody := "{\"name\":\"test\",\"channels\":[\"test_channel\"],\"data\":\"yolo\"}"
		actualBody, err := ioutil.ReadAll(req.Body)
		assert.Equal(t, expectedBody, string(actualBody))
		assert.Equal(t, "application/json", req.Header["Content-Type"][0])
		assert.NoError(t, err)

	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	client := Client{AppId: "id", Key: "key", Secret: "secret", Host: u.Host}
	err := client.Trigger("test_channel", "test", "yolo")
	assert.NoError(t, err)
}

func TestGetChannelsSuccessCase(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		testJSON := "{\"channels\":{\"presence-session-d41a439c438a100756f5-4bf35003e819bb138249-5cbTiUiPNGI\":{\"user_count\":1},\"presence-session-d41a439c438a100756f5-4bf35003e819bb138249-PbZ5E1pP8uF\":{\"user_count\":1},\"presence-session-d41a439c438a100756f5-4bf35003e819bb138249-oz6iqpSxMwG\":{\"user_count\":1}}}"

		fmt.Fprintf(res, testJSON)
		assert.Equal(t, "GET", req.Method)

	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	client := Client{AppId: "id", Key: "key", Secret: "secret", Host: u.Host}
	channels, err := client.Channels(nil)
	assert.NoError(t, err)

	expected := &ChannelsList{
		Channels: map[string]ChannelListItem{
			"presence-session-d41a439c438a100756f5-4bf35003e819bb138249-5cbTiUiPNGI": ChannelListItem{UserCount: 1},
			"presence-session-d41a439c438a100756f5-4bf35003e819bb138249-PbZ5E1pP8uF": ChannelListItem{UserCount: 1},
			"presence-session-d41a439c438a100756f5-4bf35003e819bb138249-oz6iqpSxMwG": ChannelListItem{UserCount: 1},
		},
	}

	assert.Equal(t, channels, expected)
}

func TestGetChannelSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		testJSON := "{\"user_count\":1,\"occupied\":true,\"subscription_count\":1}"

		fmt.Fprintf(res, testJSON)
		assert.Equal(t, "GET", req.Method)

	}))

	defer server.Close()
	u, _ := url.Parse(server.URL)
	client := Client{AppId: "id", Key: "key", Secret: "secret", Host: u.Host}
	channel, err := client.Channel("test_channel", nil)
	assert.NoError(t, err)

	expected := &Channel{
		Name:              "test_channel",
		Occupied:          true,
		UserCount:         1,
		SubscriptionCount: 1,
	}

	assert.Equal(t, channel, expected)
}

func TestGetChannelUserSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		testJSON := "{\"users\":[{\"id\":\"red\"},{\"id\":\"blue\"}]}"

		fmt.Fprintf(res, testJSON)
		assert.Equal(t, "GET", req.Method)

	}))

	defer server.Close()
	u, _ := url.Parse(server.URL)
	client := Client{AppId: "id", Key: "key", Secret: "secret", Host: u.Host}
	users, err := client.GetChannelUsers("test_channel")
	assert.NoError(t, err)

	expected := &Users{
		List: []User{User{Id: "red"}, User{Id: "blue"}},
	}

	assert.Equal(t, users, expected)
}

func TestTriggerWithSocketId(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)

		expectedBody := "{\"name\":\"test\",\"channels\":[\"test_channel\"],\"data\":\"yolo\",\"socket_id\":\"1234.12\"}"
		actualBody, err := ioutil.ReadAll(req.Body)
		assert.Equal(t, expectedBody, string(actualBody))
		assert.NoError(t, err)

	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	client := Client{AppId: "id", Key: "key", Secret: "secret", Host: u.Host}
	client.TriggerExclusive("test_channel", "test", "yolo", "1234.12")
}

func TestTriggerSocketIdValidation(t *testing.T) {
	client := Client{AppId: "id", Key: "key", Secret: "secret"}
	err := client.TriggerExclusive("test_channel", "test", "yolo", "1234.12:lalala")
	assert.Error(t, err)
}

func TestTriggerBatch(t *testing.T) {
	expectedBody := `{"batch":[{"channel":"test_channel","name":"test","data":"yolo1"},{"channel":"test_channel","name":"test","data":"yolo2"}]}`
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		fmt.Fprintf(res, "{}")
		assert.Equal(t, "POST", req.Method)

		actualBody, err := ioutil.ReadAll(req.Body)
		assert.Equal(t, expectedBody, string(actualBody))
		assert.Equal(t, "application/json", req.Header["Content-Type"][0])
		assert.Equal(t, "/apps/appid/batch_events", req.URL.Path)
		assert.NoError(t, err)

	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	client := Client{AppId: "appid", Key: "key", Secret: "secret", Host: u.Host}

	err := client.TriggerBatch([]Event{
		{"test_channel", "test", "yolo1", nil},
		{"test_channel", "test", "yolo2", nil},
	})

	assert.NoError(t, err)
}

func TestErrorResponseHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(400)
		fmt.Fprintf(res, "Cannot retrieve the user count unless the channel is a presence channel")

	}))

	defer server.Close()

	u, _ := url.Parse(server.URL)
	client := Client{AppId: "id", Key: "key", Secret: "secret", Host: u.Host}

	channelParams := map[string]string{"info": "user_count,subscription_count"}
	channel, err := client.Channel("this_is_not_a_presence_channel", channelParams)

	assert.Error(t, err)
	assert.EqualError(t, err, "Status Code: 400 - Cannot retrieve the user count unless the channel is a presence channel")
	assert.Nil(t, channel)
}

func TestRequestTimeouts(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Second * 1)
		// res.WriteHeader(200)
		fmt.Fprintf(res, "{}")
	}))

	defer server.Close()

	u, _ := url.Parse(server.URL)
	client := Client{AppId: "id", Key: "key", Secret: "secret", Host: u.Host, HttpClient: &http.Client{Timeout: time.Millisecond * 100}}

	err := client.Trigger("test_channel", "test", "yolo")

	assert.Error(t, err)

}

func TestChannelLengthValidation(t *testing.T) {
	channels := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}

	client := Client{AppId: "id", Key: "key", Secret: "secret"}
	err := client.TriggerMulti(channels, "yolo", "woot")

	assert.EqualError(t, err, "You cannot trigger on more than 10 channels at once")
}

func TestChannelFormatValidation(t *testing.T) {
	channel1 := "w000^$$£@@@"

	var channel2 string

	for i := 0; i <= 202; i++ {
		channel2 += "a"
	}

	client := Client{AppId: "id", Key: "key", Secret: "secret"}
	err1 := client.Trigger(channel1, "yolo", "w00t")

	err2 := client.Trigger(channel2, "yolo", "not 19 forever")

	assert.EqualError(t, err1, "At least one of your channels' names are invalid")

	assert.EqualError(t, err2, "At least one of your channels' names are invalid")

}

func TestDataSizeValidation(t *testing.T) {
	client := Client{AppId: "id", Key: "key", Secret: "secret"}

	var data string

	for i := 0; i <= 10242; i++ {
		data += "a"
	}
	err := client.Trigger("channel", "event", data)

	assert.EqualError(t, err, "Data must be smaller than 10kb")

}

func TestInitialisationFromURL(t *testing.T) {
	url := "http://feaf18a411d3cb9216ee:fec81108d90e1898e17a@api.pusherapp.com/apps/104060"
	client, _ := ClientFromURL(url)
	expectedClient := &Client{Key: "feaf18a411d3cb9216ee", Secret: "fec81108d90e1898e17a", AppId: "104060", Host: "api.pusherapp.com"}
	assert.Equal(t, expectedClient, client)
}

func TestURLInitErrorNoSecret(t *testing.T) {
	url := "http://fec81108d90e1898e17a@api.pusherapp.com/apps"
	client, err := ClientFromURL(url)
	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestURLInitHTTPS(t *testing.T) {
	url := "https://key:secret@api.pusherapp.com/apps/104060"
	client, _ := ClientFromURL(url)
	assert.True(t, client.Secure)
}

func TestURLInitErrorNoID(t *testing.T) {
	url := "http://fec81108d90e1898e17a@api.pusherapp.com/apps"
	client, err := ClientFromURL(url)
	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestInitialisationFromENV(t *testing.T) {
	os.Setenv("PUSHER_URL", "http://feaf18a411d3cb9216ee:fec81108d90e1898e17a@api.pusherapp.com/apps/104060")
	client, _ := ClientFromEnv("PUSHER_URL")
	expectedClient := &Client{Key: "feaf18a411d3cb9216ee", Secret: "fec81108d90e1898e17a", AppId: "104060", Host: "api.pusherapp.com"}
	assert.Equal(t, expectedClient, client)
}

func TestNotifySuccess(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(`{"number_of_subscribers": 10}`))
	}))

	defer server.Close()

	u, _ := url.Parse(server.URL)
	client := Client{AppId: "id", Key: "key", Secret: "secret", PushNotificationHost: u.Host, HttpClient: &http.Client{Timeout: time.Millisecond * 100}}

	testPN := PushNotification{
		WebhookURL: "testURL",
		GCM:        []byte(`hello`),
	}

	interests := []string{"testInterest"}
	response, err := client.Notify(interests, testPN)

	assert.Equal(t, 10, response.NumSubscribers, "returned response.NumSubscribers should be equal to the server response body amount")
	assert.NoError(t, err)
}

func TestNotifySuccessNoSubscribers(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusAccepted)
		res.Write([]byte(`{"number_of_subscribers":0}`))
	}))

	defer server.Close()

	u, _ := url.Parse(server.URL)
	client := Client{AppId: "id", Key: "key", Secret: "secret", PushNotificationHost: u.Host, HttpClient: &http.Client{Timeout: time.Millisecond * 100}}

	testPN := PushNotification{
		WebhookURL: "testURL",
		GCM:        []byte(`hello`),
	}

	interests := []string{"testInterest"}
	response, err := client.Notify(interests, testPN)

	assert.Equal(t, 0, response.NumSubscribers, "returned response.NumSubscribers should be equal to the server response body amount")
	assert.NoError(t, err)
}

func TestNotifyServerError(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusInternalServerError)
	}))

	defer server.Close()

	u, _ := url.Parse(server.URL)
	client := Client{AppId: "id", Key: "key", Secret: "secret", PushNotificationHost: u.Host, HttpClient: &http.Client{Timeout: time.Millisecond * 100}}

	testPN := PushNotification{
		WebhookURL: "testURL",
		GCM:        []byte(`hello`),
	}

	interests := []string{"testInterest"}

	response, err := client.Notify(interests, testPN)

	assert.Nil(t, response, "response should return nil on error")
	assert.Error(t, err)
	assert.EqualError(t, err, "Status Code: 500 - ")
}

func TestNotifyInvalidPushNotification(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusInternalServerError)
	}))

	defer server.Close()

	u, _ := url.Parse(server.URL)
	client := Client{AppId: "id", Key: "key", Secret: "secret", PushNotificationHost: u.Host, HttpClient: &http.Client{Timeout: time.Millisecond * 100}}

	testPN := PushNotification{
		WebhookURL: "testURL",
	}

	interests := []string{"testInterest"}

	response, err := client.Notify(interests, testPN)

	assert.Nil(t, response, "response should return nil on error")
	assert.Error(t, err)
}

func TestNotifyNoPushNotificationHost(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusInternalServerError)
	}))

	defer server.Close()

	client := Client{AppId: "id", Key: "key", Secret: "secret", HttpClient: &http.Client{Timeout: time.Millisecond * 100}}

	testPN := PushNotification{
		WebhookURL: "testURL",
	}

	interests := []string{"testInterest"}

	response, err := client.Notify(interests, testPN)

	assert.Nil(t, response, "response should return nil on error")
	assert.Error(t, err)
}
