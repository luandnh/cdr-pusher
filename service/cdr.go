package service

import (
	"cdr-pusher/common/log"
	"cdr-pusher/common/model"
	"cdr-pusher/internal/redis"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/clbanning/mxj/v2"
	"github.com/go-resty/resty/v2"
)

type (
	ICdr interface {
		HandlePostXmlToAPI(cdr []byte) error
		HandlePushCdr()
	}
	Cdr struct {
		APICdrUrl string
	}
)

func NewCdr(apiCdrUrl string) ICdr {
	return &Cdr{
		APICdrUrl: apiCdrUrl,
	}
}

const (
	CDR_FAIL_LIST = "cdr_fail_list"
	CDR_FAIL_DIR  = "tmp/fail/"
)

var ctxBg = context.Background()

func (s *Cdr) HandlePushCdr() {
	var listUuidFail []string
	listCdr, err := redis.Redis.HGetAll(CDR_FAIL_LIST)
	if err != nil {
		log.Error("Get list cdr failed: ", err)
		return
	}
	if len(listCdr) < 1 {
		return
	}
	for uuid, value := range listCdr {
		cdrBytes, err := s.readCdrFromFile(uuid)
		if err != nil {
			log.Error("Get cdr failed: ", err)
			continue
		} else if cdrBytes == nil {
			log.Error("Get cdr failed: nil")
			failCount, err := s.HandleUpdateCdrRedis(uuid, value)
			log.Error("Fail count : ", failCount)
			log.Error("Loi ne", err)
			if err != nil {
				log.Error("Update CDR err : ", err)
			}
			if failCount > 5 {
				listUuidFail = append(listUuidFail, string("Uuid "+uuid+" => fail \r\n"))
			}

			continue
		}
		if err := s.HandlePostXmlToAPI(cdrBytes); err != nil {
			log.Error("Post cdr failed err : ", err)
			failCount, err := s.HandleUpdateCdrRedis(uuid, value)
			log.Error("Fail count : ", failCount)
			log.Error("Loi ne", err)
			if err != nil {
				log.Error("Update CDR err : ", err)
			}
			if failCount > 5 {
				listUuidFail = append(listUuidFail, string("Uuid "+uuid+" => fail \r\n"))
			}
			// if failCount, err := s.HandleUpdateCdrRedis(uuid, value); err != nil {
			// 	log.Error("Update CDR err : ", err)
			// 	log.Error("Fail count: ", fail_count)
			// }
			continue
		} else {
			if err := redis.Redis.HDel(CDR_FAIL_LIST, uuid); err != nil {
				log.Error("HMDel CDR err : ", err)
			} else {
				if err := s.delCdrFile(uuid); err != nil {
					log.Error("Del CDR err : ", err)
				}
			}

		}
	}
	var msg string
	if len(listUuidFail) > 0 {
		for _, value := range listUuidFail {
			msg += value
		}
		log.Info("Msg ", msg)
		if msg != "" {
			err := sendMail(msg)
			if err != nil {
				log.Info("Err send mail ", err)
			}
			listUuidFail = nil
		}
	}
}

func (s *Cdr) HandleUpdateCdrRedis(uuid string, value string) (int, error) {
	cdrRedis := new(model.CdrRedis)
	if len(uuid) < 1 {
		return cdrRedis.FailedCount, errors.New("uuid is nil")
	}

	if len(value) > 0 {
		if err := json.Unmarshal([]byte(value), cdrRedis); err != nil {
			log.Error("Unmarshal CDR err : ", err)
			return cdrRedis.FailedCount, err
		}
	}
	cdrRedis.FailedCount += 1
	cdrRedis.LastPushedAt = time.Now().Local().Format("2006-01-02 15:04:05")
	val, err := json.Marshal(cdrRedis)
	if err != nil {
		log.Error("Marshal CDR err : ", err)
		return cdrRedis.FailedCount, err
	}
	data := []interface{}{uuid, string(val)}
	if _, err := redis.Redis.HSet(CDR_FAIL_LIST, data); err != nil {
		log.Error("HSet CDR err : ", err)
		return cdrRedis.FailedCount, err
	}
	return cdrRedis.FailedCount, nil
}

func (s *Cdr) HandlePostXmlToAPI(cdr []byte) error {
	cdrUuid := ""
	mv, err := mxj.NewMapXml(cdr)
	if err != nil {
		log.Error("Body to Map err: ", err)
		return err
	} else {
		variables, err := mv.ValueForKey("variables")
		if err != nil {
			log.Error("Get XML Uuid err: ", err)
		}
		variablesMap, _ := variables.(map[string]interface{})
		cdrUuid, _ = variablesMap["uuid"].(string)
		log.Info(fmt.Sprintf("Push CDR uuid %s", cdrUuid))
	}
	client := resty.New()
	client.SetTimeout(time.Second * 3)
	client.SetTLSClientConfig(&tls.Config{
		InsecureSkipVerify: true,
	})
	res, err := client.R().
		SetHeader("Content-Type", "application/xml").
		SetBody(string(cdr)).
		Post(s.APICdrUrl)
	if err != nil {
		log.Error("Post Cdr Xml : ", err)
		if fail_count, err := s.HandleUpdateCdrRedis(cdrUuid, ""); err != nil {
			log.Error("Update CDR err : ", err)
			log.Error("Fail count: ", fail_count)
		}
		if err := s.saveCdrToFile(cdrUuid, cdr); err != nil {
			log.Error("Write CDR err : ", err)
		}
		return err
	} else if (res.StatusCode() != http.StatusCreated) && (res.StatusCode() != http.StatusOK) {
		if res.StatusCode() == http.StatusUnprocessableEntity {
			if err := redis.Redis.HDel(CDR_FAIL_LIST, cdrUuid); err != nil {
				log.Error("HMDel CDR err : ", err)
			}
		} else {
			if fail_count, err := s.HandleUpdateCdrRedis(cdrUuid, ""); err != nil {
				log.Error("Update CDR err : ", err)
				log.Error("Fail count: ", fail_count)
			}
			if err := s.saveCdrToFile(cdrUuid, cdr); err != nil {
				log.Error("Write CDR err : ", err)
			}
		}
		return errors.New("post fail")
	} else {
		return nil
	}

}

func (s *Cdr) saveCdrToFile(uuid string, value []byte) error {
	if _, err := os.Stat(CDR_FAIL_DIR); os.IsNotExist(err) {
		_ = os.MkdirAll(CDR_FAIL_DIR, 0755)
	}
	f, err := os.Create(CDR_FAIL_DIR + uuid)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(value); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *Cdr) readCdrFromFile(uuid string) ([]byte, error) {
	return os.ReadFile(CDR_FAIL_DIR + uuid)
}

func (s *Cdr) delCdrFile(uuid string) error {
	return os.Remove(CDR_FAIL_DIR + uuid)
}

func sendMail(msg string) error {
	addMail := "noreply@tel4vn.com"
	passMail := "klcszlredhdlsmww"
	hostMail := "smtp.gmail.com"
	portMail := "587"
	toList := []string{"tuananh@tel4vn.com"}
	errMsg := []byte("Subject: CDR Notification!\r\n" +
		"\r\n" +
		msg + ".\r\n")
	auth := smtp.PlainAuth("", addMail, passMail, hostMail)
	err := smtp.SendMail(hostMail+":"+portMail, auth, addMail, toList, errMsg)
	if err != nil {
		log.Error("Send mail err: ", err.Error())
		return err
	}
	log.Info("Send mail success")
	return nil
}
