package tasks

import (
	"fmt"
	"goquery"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (t *Task) Watch() error {
	headers := [][2]string{
		{"accept", "*/*"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/x-www-form-urlencoded"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"},
	}

	values := url.Values{
		"term":                  {t.Term},
		"courseReferenceNumber": {t.Crns},
	}

	response, err := t.DoReq(t.MakeReq("POST", "https://reg-prod.ec.fhda.edu/StudentRegistrationSsb/ssb/searchResults/getEnrollmentInfo", headers, []byte(values.Encode())))
	if err != nil {
		fmt.Println(err)
		discardResp(response)
	}

	body, _ := readBody(response)
	reader := strings.NewReader(string(body))
	document, _ := goquery.NewDocumentFromReader(reader)
	var enrollmentSeatsAvailable, waitlistCapacity, waitlistActual, waitlistSeatsAvailable string

	document.Find("span.status-bold").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "Enrollment Seats Available:") {
			enrollmentSeatsAvailable = s.Next().Text()
		} else if strings.Contains(s.Text(), "Waitlist Seats Available:") {
			waitlistSeatsAvailable = s.Next().Text()
		} else if strings.Contains(s.Text(), "Waitlist Capacity:") {
			waitlistCapacity = s.Next().Text()
		} else if strings.Contains(s.Text(), "Waitlist Actual:") {
			waitlistActual = s.Next().Text()
		}
	})

	numEnrollmentSeatsAvailable, _ := strconv.Atoi(enrollmentSeatsAvailable)
	numWaitlistCapacity, _ := strconv.Atoi(waitlistCapacity)
	numWaitlistActual, _ := strconv.Atoi(waitlistActual)
	numWaitlistSeatsAvailable, _ := strconv.Atoi(waitlistSeatsAvailable)

	if numWaitlistCapacity > numWaitlistActual && (numWaitlistSeatsAvailable > 0) || (numEnrollmentSeatsAvailable > 0 && numWaitlistSeatsAvailable > 0) {
		t.Status = "Now available"
	} else {
		if numEnrollmentSeatsAvailable >= 1 && numWaitlistSeatsAvailable == 0 {
			t.Status = "Waitlist opening soon"
		} else {
			t.Status = "Not available"
		}
		time.Sleep(1 * time.Second)
		return t.Watch()
	}
	return nil
}
