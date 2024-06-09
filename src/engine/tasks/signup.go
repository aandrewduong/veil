package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type SignupSession struct {
	SAMLRequest string
	Model       map[string]interface{}
}

func (t *Task) GetRegistrationStatus() error {
	t.Status = "Getting Registration Status"
	headers := [][2]string{
		{"accept", "*/*"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/x-www-form-urlencoded"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"},
	}

	values := url.Values{
		"term":            {t.Term},
		"studyPath":       {},
		"startDatepicker": {},
		"endDatepicker":   {},
		"uniqueSessionId": {t.Session.UniqueSessionId},
	}

	response, err := t.DoReq(t.MakeReq("POST", "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/term/search?mode=registration", headers, []byte(values.Encode())))
	if err != nil {
		discardResp(response)
		return err
	}

	body, _ := readBody(response)
	fmt.Println(string(body))

	var registrationStatus RegistrationStatus
	if err := json.Unmarshal(body, &registrationStatus); err != nil {
		return err
	}

	var hasFailure, hasRegistrationTime bool
	var timeFailure string

	for _, failure := range registrationStatus.StudentEligFailures {
		fmt.Println(failure)
		hasFailure = true
		if strings.Contains(failure, "You can register from") {
			hasRegistrationTime = true
			timeFailure = failure
			break
		}
	}

	if hasFailure && !hasRegistrationTime {
		return errors.New(registrationStatus.StudentEligFailures[len(registrationStatus.StudentEligFailures)-1])
	}

	if !hasFailure && !hasRegistrationTime {
		pattern := regexp.MustCompile(`\d{2}/\d{2}/\d{4} \d{2}:\d{2} [APM]{2}`)
		matches := pattern.FindAllString(timeFailure, -1)

		if len(matches) > 0 {
			location, _ := time.LoadLocation("America/Los_Angeles")
			targetTime, _ := time.ParseInLocation("01/02/2006 03:04 PM", matches[0], location)
			now := time.Now().In(location)

			if now.After(targetTime) {
				time.Sleep(2 * time.Second)
				return t.GetRegistrationStatus()
			} else if now.Before(targetTime) {
				timeToWait := targetTime.Sub(now)

				t.Status = fmt.Sprintf("Waiting til %s", targetTime.Format(time.RFC1123))
				fmt.Printf("Waiting for Registration to open: %s\n", targetTime.Format(time.RFC1123))
				fmt.Printf("Will continue in %s\n", formatDuration(timeToWait))

				go func() {
					ticker := time.NewTicker(5 * time.Minute)
					defer ticker.Stop()

					endTime := time.Now().Add(timeToWait)
					for now := range ticker.C {
						if now.After(endTime) {
							break
						}
					}
				}()

				time.Sleep(timeToWait)
				return t.GetRegistrationStatus()
			}
		}
	}
	return nil
}

func (t *Task) VisitClassRegistration() error {
	t.Status = "Visiting Class Registration"

	headers := [][2]string{
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		{"accept-language", "en-US,en;q=0.9"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"},
	}

	response, err := t.DoReq(t.MakeReq("HEAD", "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/classRegistration/classRegistration", headers, nil))
	if err != nil {
		discardResp(response)
		return err
	}
	return nil
}

func (t *Task) AddCourse(course string) error {
	t.Status = "Adding Course"

	headers := [][2]string{
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		{"accept-language", "en-US,en;q=0.9"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"},
	}

	url := fmt.Sprintf("https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/classRegistration/addRegistrationItem?term=%s&courseReferenceNumber=%s&olr=false", t.Term, course)
	response, err := t.DoReq(t.MakeReq("GET", url, headers, nil))
	if err != nil {
		discardResp(response)
		return err
	}

	body, _ := readBody(response)
	fmt.Println(string(body))

	var addCourse AddCourse
	if err := json.Unmarshal(body, &addCourse); err != nil {
		return err
	}

	if addCourse.Success {
		model, err := extractModel([]byte(body))
		if err != nil {
			return err
		}
		model["selectedAction"] = "WL"
		t.Session.SignupSession.Model = model
	} else {
		t.Status = addCourse.Message
	}
	return nil
}

func (t *Task) AddCourses() error {
	for _, course := range t.CRNs {
		if err := t.AddCourse(course); err != nil {
			return err
		}
	}
	return nil
}

func (t *Task) SendBatch() error {
	t.Status = "Submitting Batch"

	headers := [][2]string{
		{"accept", "application/json"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/json"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"},
	}

	batch := Batch{
		Update:          []map[string]interface{}{t.Session.SignupSession.Model},
		UniqueSessionId: t.Session.UniqueSessionId,
	}

	batchJson, err := json.MarshalIndent(batch, "", "  ")
	if err != nil {
		return err
	}

	response, err := t.DoReq(t.MakeReq("POST", "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/classRegistration/submitRegistration/batch", headers, []byte(batchJson)))
	if err != nil {
		discardResp(response)
		return err
	}

	body, _ := readBody(response)
	fmt.Println(string(body))

	var changes Changes
	if err := json.Unmarshal(body, &changes); err != nil {
		return err
	}

	for _, data := range changes.Data.Update {
		for _, courseReferenceNumber := range t.CRNs {
			if data.CourseReferenceNumber == courseReferenceNumber {
				switch data.StatusDescription {
				case "Registered":
					t.Status = "Registered"
					t.SendNotification(data.CourseTitle, "Registered")
				case "Waitlisted":
					t.Status = "Waitlisted"
					t.SendNotification(data.CourseTitle, "Waitlisted")
				case "Errors Preventing Registration":
					t.Status = data.CrnErrors[0].Message
					t.SendNotification(data.CourseTitle, data.CrnErrors[0].Message)
				}
			}
		}
	}
	return nil
}

func (t *Task) Signup() {
	t.HomepageURL = "https://ssb-prod.ec.fhda.edu/ssomanager/saml/login?relayState=%2Fc%2Fauth%2FSSB%3Fpkg%3Dhttps%3A%2F%2Fssb-prod.ec.fhda.edu%2FPROD%2Ffhda_uportal.P_DeepLink_Post%3Fp_page%3Dbwskfreg.P_AltPin%26p_payload%3De30%3D"
	t.SSOManagerURL = "https://ssb-prod.ec.fhda.edu/ssomanager/saml/SSO"
	t.GenSession()
	t.GetRegistrationStatus()
	t.VisitClassRegistration()
	t.AddCourses()
	t.SendBatch()
}
