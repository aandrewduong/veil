package tasks

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"goquery"
)

type Session struct {
	LoginAttempts   int
	SAMLResponse    string
	RelayState      string
	SignupSession   SignupSession
	UniqueSessionId string
}

// GenSessionId generates a unique session ID for the task.
func (t *Task) GenSessionId() error {
	t.Session.UniqueSessionId = fmt.Sprintf("%s%v", strings.ToLower(generateRandomString(5)), time.Now().UnixNano()/int64(time.Millisecond))
	return nil
}

// VisitHomepage sends a GET request to the homepage URL.
func (t *Task) VisitHomepage() error {
	t.Status = "Visiting Homepage"
	headers := [][2]string{
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		{"accept-language", "en-US,en;q=0.9"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"},
	}

	response, err := t.DoReq(t.MakeReq("GET", t.HomepageURL, headers, nil))
	if err != nil {
		discardResp(response)
		return err
	}
	return nil
}

// Login sends a login request with the provided username and password.
func (t *Task) Login() error {
	t.Status = "Logging In"
	headers := [][2]string{
		{"accept", "*/*"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/x-www-form-urlencoded"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"},
	}

	t.Session.LoginAttempts++

	values := url.Values{}
	values.Set("j_username", t.Username)
	values.Set("j_password", t.Password)
	values.Set("_eventId_proceed", "")

	loginURL := fmt.Sprintf("https://ssoshib.fhda.edu/idp/profile/SAML2/Redirect/SSO?execution=e1s%d", t.Session.LoginAttempts)
	response, err := t.DoReq(t.MakeReq("POST", loginURL, headers, []byte(values.Encode())))
	if err != nil {
		discardResp(response)
		return err
	}

	body, _ := readBody(response)
	reader := strings.NewReader(string(body))
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		discardResp(response)
		return err
	}

	var message string
	document.Find("div[class='alert alert-danger']").Each(func(index int, element *goquery.Selection) {
		message = strings.TrimSpace(element.Text())
	})

	switch message {
	case "The username you entered cannot be identified.":
		t.Status = "Invalid Username"
	case "The password you entered was incorrect.":
		t.Status = "Invalid Password"
	case "You may be seeing this page because you used the Back button while browsing a secure web site or application...":
		t.Status = "Bad Session"
	case "":
		break
	default:
		t.Status = message
		time.Sleep(2 * time.Second)
		return t.Login()
	}

	t.Session.RelayState = getSelectorAttr(document, "input[name='RelayState']", "value")
	t.Session.SAMLResponse = getSelectorAttr(document, "input[name='SAMLResponse']", "value")
	return nil
}

// SubmitCommonAuth sends a POST request to submit common authentication data.
func (t *Task) SubmitCommonAuth() error {
	t.Status = "Submitting Common Auth"

	headers := [][2]string{
		{"accept", "*/*"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/x-www-form-urlencoded"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"},
	}

	values := url.Values{
		"RelayState":   {t.Session.RelayState},
		"SAMLResponse": {t.Session.SAMLResponse},
	}

	response, err := t.DoReq(t.MakeReq("POST", "https://eis-prod.ec.fhda.edu/commonauth", headers, []byte(values.Encode())))
	if err != nil {
		discardResp(response)
		return err
	}

	body, _ := readBody(response)
	reader := strings.NewReader(string(body))
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		discardResp(response)
		return err
	}

	var message string
	document.Find("div[class='retry-msg-text text_right_custom']").Each(func(index int, element *goquery.Selection) {
		message = strings.TrimSpace(element.Text())
	})

	if strings.Contains(message, "Authentication Error!") {
		fmt.Println("Authentication Error!")
	}

	t.Session.RelayState = getSelectorAttr(document, "input[name='RelayState']", "value")
	t.Session.SAMLResponse = getSelectorAttr(document, "input[name='SAMLResponse']", "value")
	return nil
}

// SubmitSSOManager sends a POST request to the SSO Manager URL.
func (t *Task) SubmitSSOManager() error {
	t.Status = "Submitting SSO Manager"

	headers := [][2]string{
		{"accept", "*/*"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/x-www-form-urlencoded"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"},
	}

	values := url.Values{
		"RelayState":   {t.Session.RelayState},
		"SAMLResponse": {t.Session.SAMLResponse},
	}

	response, err := t.DoReq(t.MakeReq("POST", t.SSOManagerURL, headers, []byte(values.Encode())))
	if err != nil {
		discardResp(response)
		return err
	}
	return nil
}

// RegisterPostSignIn sends a GET request to register post sign-in.
func (t *Task) RegisterPostSignIn() error {
	t.Status = "Posting Register Sign-in"

	headers := [][2]string{
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		{"accept-language", "en-US,en;q=0.9"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"},
	}

	url := "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/registration/registerPostSignIn?mode=registration"
	response, err := t.DoReq(t.MakeReq("GET", url, headers, nil))
	if err != nil {
		discardResp(response)
		return err
	}

	body, _ := readBody(response)
	reader := strings.NewReader(string(body))
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		discardResp(response)
		return err
	}

	t.Session.SignupSession.SAMLRequest = getSelectorAttr(document, "input[name='SAMLRequest']", "value")
	return nil
}

// SubmitSamIsso sends a POST request to submit SAML SSO.
func (t *Task) SubmitSamIsso() error {
	t.Status = "Submitting Sam Isso"

	headers := [][2]string{
		{"accept", "*/*"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/x-www-form-urlencoded"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"},
	}

	values := url.Values{
		"SAMLRequest": {t.Session.SignupSession.SAMLRequest},
	}

	response, err := t.DoReq(t.MakeReq("POST", "https://eis-prod.ec.fhda.edu/samlsso", headers, []byte(values.Encode())))
	if err != nil {
		discardResp(response)
		return err
	}

	body, _ := readBody(response)
	reader := strings.NewReader(string(body))
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		discardResp(response)
		return err
	}

	t.Session.SAMLResponse = getSelectorAttr(document, "input[name='SAMLResponse']", "value")
	return nil
}

// SubmitSSBSp sends a POST request to submit SAML response to the SSB service provider.
func (t *Task) SubmitSSBSp() error {
	t.Status = "Submitting SSB SSP"

	headers := [][2]string{
		{"accept", "*/*"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/x-www-form-urlencoded"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"},
	}

	values := url.Values{
		"SAMLResponse": {t.Session.SAMLResponse},
	}

	url := "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/saml/SSO/alias/registrationssb-prod-sp"
	response, err := t.DoReq(t.MakeReq("POST", url, headers, []byte(values.Encode())))
	if err != nil {
		discardResp(response)
		return err
	}
	return nil
}

// GenSession generates a new session and performs all the required steps.
func (t *Task) GenSession() {
	t.GenSessionId()
	t.VisitHomepage()
	t.Login()
	t.SubmitCommonAuth()
	t.SubmitSSOManager()
	t.RegisterPostSignIn()
	t.SubmitSamIsso()
	t.SubmitSSBSp()
}
