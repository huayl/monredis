package agent

import (
	"fmt"
	"strings"
	"net/http"
	"io/ioutil"
	
	"github.com/garyburd/redigo/redis"
)


func httpPost(postData string) error{
	Debug("http post|%s|%s", Settings.HaproxyAddr, postData)
	if len(postData) < 5 {
		Error("post data is invalid:%s", postData)
		return nil
	}
	
    resp, err := http.Post(Settings.HaproxyAddr, "application/x-www-form-urlencoded", strings.NewReader(postData))
    if err != nil {
		Error("post err:%v", err)
		return err
    }
 
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        // handle error
		Error("http get err:%v|%v", body, err)
		return err
    }

    //Debug(string(body))
	return nil
}


func ComposeRedisAddress(ip string, port string) (address string) {
	address = fmt.Sprint(ip, ":", port)
	return address
}


func UpdateMaster(master_name string, ip string, port string, disable bool) error {
	address := ComposeRedisAddress(ip, port)
	Debug(fmt.Sprintf("Updating master %s to %s.", master_name, address))
	var httpData string
	
	if disable {
		httpData = fmt.Sprintf("s=%s%s%s%s", master_name, "%3A", address, "&action=disable&b=%234")
	}else {
		httpData = fmt.Sprintf("s=%s%s%s%s", master_name, "%3A", address, "&action=enable&b=%234")
	}
	
	return httpPost(httpData)
}


func GetSentinel() (sentinel string) {
	address := ComposeRedisAddress(Settings.SentinelIp, Settings.SentinelPort)
	return address
}

func SwitchMaster(master_name string, ip string, port string) error {
	Debug("Received switch-master.")
	address := ComposeRedisAddress(ip, port)
	
	httpData := fmt.Sprintf("s=%s%s%s%s", master_name, "%3A", address, "&action=enable&b=%234")
	return httpPost(httpData)
}

func ValidateCurrentMaster() error {
	c, err := redis.Dial("tcp", GetSentinel())
	if err != nil {
		Error("Dial sentinel addr err|%v", err)
		return err
	}

	reply, err := redis.Values(c.Do("SENTINEL", "masters"))
	if err != nil {
		Error("Call sentinel cmd err|%v", err)
		return err
	}

	var sentinel_info []string

	reply, err = redis.Scan(reply, &sentinel_info)
	if err != nil {
		Error("Parse sentinel result err|%v", err)
		return err
	}
	master_name := sentinel_info[1]
	ip          := sentinel_info[3]
	port        := sentinel_info[5]

	err = SwitchMaster(master_name, ip, port)

	return err
}

func SubscribeToSentinel() {
	sentinel := GetSentinel()
	c, err := redis.Dial("tcp", sentinel)
	if err != nil {
		Fatal("Cannot connect to redis sentinel:", sentinel)
	}

	err = ValidateCurrentMaster()
	if err != nil {
		Fatal("Cannot switch to current master")
	}
	psc := redis.PubSubConn{c}
	Debug("Subscribing to sentinel (+switch-master).")
	psc.Subscribe("+switch-master")
	psc.Subscribe("+sdown")
	psc.Subscribe("+odown")
	psc.Subscribe("-odown")
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			Debug(fmt.Sprintf("%s: message: %s", v.Channel, v.Data))
			data := strings.Split(string(v.Data), string(' '))
			if v.Channel == "+switch-master" {
				go SwitchMaster(data[0], data[3], data[4])
			    go UpdateMaster(data[0], data[1], data[2], true)
			} else if v.Channel == "+sdown" && data[0] == "slave"{
				go UpdateMaster(data[5], data[2], data[3], true)
			}else if v.Channel == "+odown" && data[0] == "master"{
				go UpdateMaster(data[1], data[2], data[3], true)
			}else if v.Channel == "-odown" && data[0] == "master"{
				go UpdateMaster(data[1], data[2], data[3], false)
			}else{
				Error("Error channenl info>> %s: message: %s", v.Channel, v.Data)
			}

		case redis.Subscription:
			Debug(fmt.Sprintf("%s: %s %d", v.Channel, v.Kind, v.Count))
		case error:
			Fatal("Error with redis connection:", psc)
		}
	}
}

func Run() {
	Info("monitor is actived")
	SubscribeToSentinel()
}
