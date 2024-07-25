package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const version = "1.4"

type UpdateResponse struct {
	Version string `json:"version"`
}

type KeyConfig struct {
	APIKey string `json:"api_key"`
}

type UserResponse struct {
	Data struct {
		Points int    `json:"Points"`
		ID     int    `json:"ID"`
		Name   string `json:"Name"`
	} `json:"data"`
}

type SignInResponse struct {
	Data string `json:"data"`
	Code int    `json:"code"`
}

type WithdrawResponse struct {
	Data struct {
		Records []struct {
			ID      int     `json:"id"`
			Account string  `json:"account"`
			Target  string  `json:"target"`
			Points  int     `json:"points"`
			Money   float64 `json:"money"`
			Status  string  `json:"status"`
		} `json:"Records"`
	} `json:"data"`
}

func main() {
	// 检测更新
	/*updateURL := "http://api.v2.imxingkong.top:8000/update/"
	resp, err := http.Get(updateURL)
	if err != nil {
		fmt.Printf("连接服务器失败，当前版本%s\n", version)
	} else {
		defer resp.Body.Close()
		var update UpdateResponse
		json.NewDecoder(resp.Body).Decode(&update)
		if update.Version != version {
			fmt.Printf("有新版本：%s, 当前版本%s\n", update.Version, version)
		} else {
			fmt.Printf("已是最新版，当前版本%s\n", version)
		}
	}
	fmt.Println("==============================")*/

	// 检测并创建API默认文件夹
	configDir := "config"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.Mkdir(configDir, os.ModePerm)
		fmt.Println("配置文件夹已创建，请修改config目录里的key.json")
	}

	// 检测并创建API默认文件
	keyFilePath := filepath.Join(configDir, "key.json")
	if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
		defaultConfig := KeyConfig{APIKey: ""}
		configFile, _ := os.Create(keyFilePath)
		defer configFile.Close()
		json.NewEncoder(configFile).Encode(defaultConfig)
		fmt.Println("已创建key.json文件，请在其中填写API密钥")
	}

	// 读取API信息
	keyFile, err := os.Open(keyFilePath)
	if err != nil {
		fmt.Println("读取key.json文件失败")
		return
	}
	defer keyFile.Close()
	var keyConfig KeyConfig
	json.NewDecoder(keyFile).Decode(&keyConfig)

	if keyConfig.APIKey == "" {
		fmt.Println("请填写API密钥")
		fmt.Println("Powered by xingkongqwq")
		return
	}

	// 请求用户信息
	userURL := "https://api.v2.rainyun.com/user/"
	req, _ := http.NewRequest("GET", userURL, nil)
	req.Header.Add("X-Api-Key", keyConfig.APIKey)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("获取用户信息失败")
		return
	}
	defer res.Body.Close()
	var userRes UserResponse
	json.NewDecoder(res.Body).Decode(&userRes)

	points := userRes.Data.Points
	id := userRes.Data.ID
	name := userRes.Data.Name
	fmt.Printf("ID：%d\n用户名：%s\n剩余积分：%d\n", id, name, points)
	fmt.Println("==============================")

	// 签到部分
	signInURL := "https://api.v2.rainyun.com/user/reward/tasks"
	signInPayload := bytes.NewBuffer([]byte(`{"task_name": "每日签到", "verifyCode": ""}`))
	req, _ = http.NewRequest("POST", signInURL, signInPayload)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("X-Api-Key", keyConfig.APIKey)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("签到请求失败")
		return
	}
	defer res.Body.Close()
	var signInRes SignInResponse
	json.NewDecoder(res.Body).Decode(&signInRes)

	if signInRes.Data == "ok" {
		fmt.Printf("签到成功，当前剩余积分：%d\n", points+300)
	} else if signInRes.Code == 30011 {
		fmt.Println("签到失败")
	}
	fmt.Println("==============================")

	// 自动申请提现部分
	withdrawURL := "https://api.v2.rainyun.com/user/reward/withdraw"
	withdrawPayload := bytes.NewBuffer([]byte(fmt.Sprintf(`{"points": %d, "target": "%s"}`, points, keyConfig.APIKey)))
	req, _ = http.NewRequest("POST", withdrawURL, withdrawPayload)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("X-Api-Key", keyConfig.APIKey)
	if points >= 60000 {
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("自动提现请求失败")
			return
		}
		defer res.Body.Close()
		fmt.Println("自动提现成功")
	} else {
		fmt.Printf("自动提现失败，当前积分：%d\n", points)
	}
	fmt.Println("==============================")

	// 提现列表部分
	options := `{"columnFilters":{},"sort":[],"page":1,"perPage":20}`
	withdrawListURL := "https://api.v2.rainyun.com/user/reward/withdraw?options=" + options
	req, _ = http.NewRequest("GET", withdrawListURL, nil)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("X-Api-Key", keyConfig.APIKey)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("获取提现记录失败")
		return
	}
	defer res.Body.Close()

	var withdrawRes WithdrawResponse
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&withdrawRes); err != nil {
		fmt.Println("JSON解析失败")
		return
	}

	if len(withdrawRes.Data.Records) > 0 {
		record := withdrawRes.Data.Records[0]
		fmt.Println("提现记录：")
		fmt.Printf("提现ID：%d\n", record.ID)
		fmt.Printf("提现账户：%s\n", record.Account)
		fmt.Printf("提现方式：%s\n", record.Target)
		fmt.Printf("提现积分：%d\n", record.Points)
		fmt.Printf("提现金额：%.2f\n", record.Money) // 修改为浮点数类型并格式化输出
		fmt.Printf("提现状态：%s\n", record.Status)
	} else {
		fmt.Println("没有提现记录")
	}

	// 其它
	fmt.Println("Powered by xingkongqwq")

	// 脚本运行完10秒后自动关闭
	time.Sleep(10 * time.Second)
}
