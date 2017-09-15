package message

import (
	"time"

	"github.com/juju/errors"
	"github.com/zssky/log"
	"github.com/zssky/tc/cityhash"
	"github.com/zssky/tc/http"
)

type SMSBaseInfo struct {
	SenderNum string `json:"senderNum"`
	Extension string `json:"extension"`
}

type MobileInfo struct {
	MobileNum string `json:"mobileNum"`
	Signature int64  `json:"signature"`
}

type SMS struct {
	SMSBaseInfo SMSBaseInfo  `json:"smsBaseInfo"`
	MobileNums  []MobileInfo `json:"mobileNums"`
	MsgContent  string       `json:"msgContent"`
}

func generateSignature(s *SMS) (*SMS, error) {

	sms := s
	for index, m := range sms.MobileNums {
		key := sms.SMSBaseInfo.SenderNum + sms.SMSBaseInfo.Extension + m.MobileNum + sms.MsgContent

		sign, err := cityhash.CityHash64([]byte(key), int64(len(key)))
		if err != nil {
			return nil, errors.Trace(err)
		}

		sms.MobileNums[index].Signature = sign
	}

	return sms, nil
}

func SendSMS(url string, s *SMS) error {

	sms, err := generateSignature(s)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("sms:%v", sms)

	data, _, err := http.PostJSON(url, []interface{}{sms}, time.Second*5, time.Second*5)
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("success, data:%v", string(data))
	return nil
}
