package service

import (
	"cdr-pusher/common/log"
	"cdr-pusher/internal/redis"
	"crypto/tls"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/clbanning/mxj/v2"
	"github.com/go-resty/resty/v2"
)

type (
	CDR struct {
		APICdrUrl string
	}
)

func NewCDR(apiCdrUrl string) CDR {
	return CDR{
		APICdrUrl: apiCdrUrl,
	}
}

var RETRY_COUNT = 100
var RETRY_WAIT_TIME = 10 * time.Second
var ALLOWED_STATUSES = []int{http.StatusCreated, http.StatusOK, http.StatusUnprocessableEntity}

const (
	R_KEY        = "fail_cdr_uuids"
	CDR_FAIL_DIR = "tmp/fail/"
)

func InArr(array interface{}, item interface{}) bool {
	arr := reflect.ValueOf(array)
	if arr.Kind() != reflect.Slice {
		log.Error("invalid slice")
		return false
	}
	for i := 0; i < arr.Len(); i++ {
		if arr.Index(i).Interface() == item {
			return true
		}
	}
	return false
}

func NewHTTPClient() *resty.Client {
	client := resty.New().
		SetTimeout(time.Second * 10).
		SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
		}).
		SetRetryCount(RETRY_COUNT).
		SetRetryWaitTime(RETRY_WAIT_TIME).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return !InArr(ALLOWED_STATUSES, r.StatusCode())
		})
	return client
}

func (s *CDR) HandlePostXmlToAPI(cdr []byte) error {
	cdrUuid := ""
	mv, err := mxj.NewMapXml(cdr)
	if err != nil {
		log.Error("xml err: ", err)
		return err
	} else {
		variables, err := mv.ValueForKey("variables")
		if err != nil {
			log.Error("xml err: ", err)
		}
		variablesMap, _ := variables.(map[string]interface{})
		cdrUuid, _ = variablesMap["uuid"].(string)
	}
	return s.pushToAPI(cdrUuid, cdr, false)
}

func (s *CDR) pushToAPI(cdrUuid string, cdr []byte, isAgain bool) error {
	client := NewHTTPClient().
		AddRetryHook(func(r *resty.Response, err error) {
			if err != nil {
				log.Error(err)
			}
			log.Warningf("retry push cdr: %s | count: %d", cdrUuid, r.Request.Attempt)
			if r.Request.Attempt == RETRY_COUNT+1 {
				addToRedis(cdrUuid, cdr)
				saveToFile(cdrUuid, cdr)
			}
		})
	r, err := client.R().
		SetHeader("Content-Type", "application/xml").
		SetBody(string(cdr)).
		Post(s.APICdrUrl)
	if err != nil {
		log.Errorf("push cdr fail: %s", cdrUuid)
		return err
	} else if InArr(ALLOWED_STATUSES, r.StatusCode()) {
		log.Infof("push cdr success: %s", cdrUuid)
		if isAgain {
			removeFromRedis(cdrUuid)
		}
	}
	return nil
}

func addToRedis(cdrUuid string, value []byte) {
	data := []interface{}{cdrUuid, string(value)}
	if _, err := redis.Redis.HSet(R_KEY, data); err != nil {
		log.Error(err)
	}
}
func removeFromRedis(cdrUuid string) {
	if err := redis.Redis.HDel(R_KEY, cdrUuid); err != nil {
		log.Error(err)
	}
}

func saveToFile(uuid string, value []byte) error {
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

func (s *CDR) HandlePushBack() {
	log.Info("start push back")
	values, err := redis.Redis.HGetAll(R_KEY)
	if err != nil {
		log.Errorf("push back err: %v", err)
		return
	}
	for k, v := range values {
		s.pushToAPI(k, []byte(v), true)
	}
	log.Info("end push back")
}
