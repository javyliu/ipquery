package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pg9182/ip2x"
)

var (
	// 数据库路径
	dbPath *string
	// API 监听地址
	listenAddr *string
	apiKey     *string
)

var (
	countryMap = make(map[string]string)
	regionMap  = make(map[string]string)
	cityMap    = make(map[string]string)
	db         *ip2x.DB
)

func init() {
	// 加载 JSON 文件
	loadMap("countries.json", &countryMap)
	loadMap("regions.json", &regionMap)
	loadMap("cities.json", &cityMap)
}

func loadMap(file string, m *map[string]string) {
	data, err := os.ReadFile(file)
	if err != nil {
		log.Printf("警告: 无法加载 %s: %v。使用空映射。", file, err)
		return
	}
	if err := json.Unmarshal(data, m); err != nil {
		log.Printf("警告: 解析 %s 失败: %v。使用空映射。", file, err)
	}
}

func localize(field, value string) string {
	// if field == "city_name" && countryCode != "CN" {
	// 	return "海外"
	// }
	switch field {
	case "country_name":
		if localized, exists := countryMap[value]; exists {
			return localized
		}
	case "region_name":
		if localized, exists := regionMap[value]; exists {
			return localized
		}
	case "city_name":
		if localized, exists := cityMap[value]; exists {
			return localized
		}
	}
	return value
}

// 查询结果结构体
type QueryResult struct {
	IP          string `json:"ip"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Region      string `json:"province"`
	City        string `json:"city"`
	Code        string `json:"code"` // 用于兼容旧api, 指明结果返回状态
	// Latitude    float32 `json:"latitude"`
	// Longitude   float32 `json:"longitude"`
	// ISP         string  `json:"isp"`
	// Domain      string  `json:"domain"`
	// ZipCode     string  `json:"zip_code"`
	// TimeZone    string  `json:"time_zone"`
	// ASN         string  `json:"asn"`
	// MobileBrand string  `json:"mobile_brand"`
	Error string `json:"error,omitempty"`
}

// api md5 验证
func md5sign(str string) string {
	t1 := md5.Sum([]byte(str))
	sign := hex.EncodeToString(t1[:])
	return sign
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// 获取参数
	params := r.URL.Query()
	ipParam := params.Get("ip")

	if ipParam == "" {
		http.Error(w, `{"error": "缺少必要参数：IP"}`, http.StatusBadRequest)
		return
	}

	// 仅在有 API Key 的情况下需要进行签名
	if *apiKey != "" {
		timeStamp := params.Get("time")
		sign := params.Get("sign")
		log.Println(timeStamp, sign)
		if timeStamp == "" || sign == "" {
			http.Error(w, `{"error": "缺少必要参数：时间戳和签名"}`, http.StatusBadRequest)
			return
		}
		needSign := md5sign(ipParam + timeStamp + *apiKey)
		log.Println(needSign)
		if sign != needSign {
			http.Error(w, `{"error": "签名不正确"}`, http.StatusBadRequest)
			return
		}
	}

	// 分割 IP 列表
	ips := strings.Split(ipParam, ",")
	if len(ips) == 0 {
		http.Error(w, `{"error": "IP 列表为空"}`, http.StatusBadRequest)
		return
	}

	// 查询结果
	results := make([]QueryResult, 0, len(ips))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			results = append(results, QueryResult{IP: ip, Error: "无效的 IP 地址"})
			continue
		}

		r, err := db.LookupString(ip)
		if err != nil {
			results = append(results, QueryResult{IP: ip, Error: fmt.Sprintf("查询失败: %v", err)})
			continue
		}

		countryCode, _ := r.GetString(ip2x.CountryCode)
		result := QueryResult{
			IP:          ip,
			Country:     localize("country_name", countryCode),
			CountryCode: countryCode,
			Region:      localize("region_name", getStringOrEmpty(r, ip2x.Region)),
			City:        localize("city_name", getStringOrEmpty(r, ip2x.City)),
			Code:        "0",
			// Latitude:    r.GetFloat32("Latitude"),
			// Longitude:   r.GetFloat32("longitude"),
			// ISP:         getStringOrEmpty(r, ip2x.ISP),
			// Domain:      getStringOrEmpty(r, ip2x.Domain),
			// ZipCode:     getStringOrEmpty(r, ip2x.ZipCode),
			// TimeZone:    getStringOrEmpty(r, ip2x.TimeZone),
			// ASN:         getStringOrEmpty(r, ip2x.ASN),
			// MobileBrand: getStringOrEmpty(r, ip2x.MobileBrand),
		}
		results = append(results, result)
	}

	// 如果仅为一个 IP 查询，返回单个结果
	if len(results) == 1 {
		result := results[0]
		if result.Error != "" {
			http.Error(w, `{"error": "查询失败"}`, http.StatusInternalServerError)
			return
		}
		// 返回 单个JSON 响应
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, `{"error": "编码响应失败"}`, http.StatusInternalServerError)
		}
		return
	}

	// 返回 JSON 响应
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, `{"error": "编码响应失败"}`, http.StatusInternalServerError)
	}
}

func getStringOrEmpty(r ip2x.Record, field ip2x.DBField) string {
	val, _ := r.GetString(field)
	return val
}

func main() {
	// 添加命令行标志以支持命令行查询（可选）
	query := flag.String("query", "", "直接查询 IP 地址（以逗号分隔）")
	dbPath = flag.String("db_path", "./IP2LOCATION-LITE-DB3.BIN", "批定数据库路径，默认为 ./IP2LOCATION-LITE-DB3.BIN")
	listenAddr = flag.String("port", ":8080", "API 监听地址，默认为 :8080")
	apiKey = flag.String("api_key", os.Getenv("IPQUERY_API_KEY"), "API Key")
	flag.Parse()

	fmt.Println("dbPath:", *dbPath)
	fmt.Println("listenAddr:", *listenAddr)

	// fmt.Println(md5sign(*query))
	// return
	// 打开数据库文件
	f, err := os.Open(*dbPath)
	if err != nil {
		log.Fatalf("错误: 无法打开数据库文件 %s: %v", *dbPath, err)
	}
	// defer f.Close()

	// 创建数据库实例
	db, err = ip2x.New(f)
	if err != nil {
		log.Fatalf("错误: 无法初始化数据库 %s: %v", *dbPath, err)
	}

	if *query != "" {
		// 命令行模式
		ips := strings.Split(*query, ",")
		results := make([]QueryResult, 0, len(ips))
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip == "" {
				results = append(results, QueryResult{IP: ip, Error: "无效的 IP 地址"})
				continue
			}

			r, _ := db.LookupString(ip)
			countryCode, _ := r.GetString(ip2x.CountryCode)
			result := QueryResult{
				IP:          ip,
				Country:     localize("country_name", countryCode),
				CountryCode: countryCode,
				Region:      localize("region_name", getStringOrEmpty(r, ip2x.Region)),
				City:        localize("city_name", getStringOrEmpty(r, ip2x.City)),
				// Latitude:    r.GetFloat32("latitude"),
				// Longitude:   r.GetFloat32("longitude"),
				// ISP:         r.GetString("isp"),
				// Domain:      r.GetString("domain"),
				// ZipCode:     r.GetString("zip_code"),
				// TimeZone:    r.GetString("time_zone"),
				// ASN:         r.GetString("asn"),
				// MobileBrand: r.GetString("mobile_brand"),
			}
			results = append(results, result)
		}

		// 打印 JSON 格式结果
		if data, err := json.MarshalIndent(results, "", "  "); err == nil {
			fmt.Println(string(data))
		} else {
			log.Fatalf("错误: 编码 JSON 失败: %v", err)
		}
		return
	}

	// API 服务模式
	http.HandleFunc("/query", queryHandler)
	log.Printf("API 服务启动在 %s，访问 /query?ip=<IP地址>", *listenAddr)
	log.Printf("API Key: %s", *apiKey)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatalf("错误: 无法启动服务: %v", err)
	}
}
