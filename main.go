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
	apiKey *string
	db     *ip2x.DB
)

var localization = struct {
	Countries map[string]string `json:"countries"`
	Regions   map[string]string `json:"regions"`
	Cities    map[string]string `json:"cities"`
}{}

func loadMap(locale string) {
	data, err := os.ReadFile(locale + ".json")
	if err != nil {
		log.Printf("警告: 无法加载 localization.json: %v。使用空映射。", err)
		return
	}
	if err := json.Unmarshal(data, &localization); err != nil {
		log.Printf("警告: 解析 localization.json 失败: %v。使用空映射。", err)
	}
}

func localize(field, value string) string {
	switch field {
	case "country_name":
		if v, ok := localization.Countries[value]; ok {
			return v
		}
	case "region_name":
		if v, ok := localization.Regions[value]; ok {
			return v
		}
	case "city_name":
		if v, ok := localization.Cities[value]; ok {
			return v
		}
	}
	return value
}

type QueryResult struct {
	IP          string `json:"ip"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Region      string `json:"province"`
	City        string `json:"city"`
	Code        string `json:"code"`
	Error       string `json:"error,omitempty"`
}

func md5sign(str string) string {
	sum := md5.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

func writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	params := r.URL.Query()
	ipParam := params.Get("ip")
	if ipParam == "" {
		writeJSONError(w, "缺少必要参数：IP", http.StatusBadRequest)
		return
	}

	if *apiKey != "" {
		timeStamp := params.Get("time")
		sign := params.Get("sign")
		if timeStamp == "" || sign == "" {
			writeJSONError(w, "缺少必要参数：时间戳和签名", http.StatusBadRequest)
			return
		}
		needSign := md5sign(ipParam + timeStamp + *apiKey)
		if sign != needSign {
			writeJSONError(w, "签名不正确", http.StatusBadRequest)
			return
		}
	}

	ips := strings.Split(ipParam, ",")
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
		results = append(results, QueryResult{
			IP:          ip,
			Country:     localize("country_name", countryCode),
			CountryCode: countryCode,
			Region:      localize("region_name", getStringOrEmpty(r, ip2x.Region)),
			City:        localize("city_name", getStringOrEmpty(r, ip2x.City)),
			Code:        "0",
		})
	}

	if len(results) == 1 {
		result := results[0]
		if result.Error != "" {
			writeJSONError(w, "查询失败", http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(result)
		return
	}
	_ = json.NewEncoder(w).Encode(results)
}

func getStringOrEmpty(r ip2x.Record, field ip2x.DBField) string {
	val, _ := r.GetString(field)
	return val
}

func runQueryMode(ips []string) {
	results := make([]QueryResult, 0, len(ips))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			results = append(results, QueryResult{IP: ip, Error: "无效的 IP 地址"})
			continue
		}
		r, _ := db.LookupString(ip)
		countryCode, _ := r.GetString(ip2x.CountryCode)
		results = append(results, QueryResult{
			IP:          ip,
			Country:     localize("country_name", countryCode),
			CountryCode: countryCode,
			Region:      localize("region_name", getStringOrEmpty(r, ip2x.Region)),
			City:        localize("city_name", getStringOrEmpty(r, ip2x.City)),
		})
	}
	if data, err := json.MarshalIndent(results, "", "  "); err == nil {
		fmt.Println(string(data))
	} else {
		log.Fatalf("错误: 编码 JSON 失败: %v", err)
	}
}

func main() {
	query := flag.String("query", "", "直接查询 IP 地址（以逗号分隔）")
	dbPath := flag.String("db_path", "./IP2LOCATION-LITE-DB3.BIN", "数据库路径")
	listenAddr := flag.String("port", ":8080", "API 监听地址")
	apiKey = flag.String("api_key", os.Getenv("IPQUERY_API_KEY"), "API Key")
	locale := flag.String("locale", "zh-CN", "本地化文件")
	flag.Parse()

	loadMap(*locale)

	f, err := os.Open(*dbPath)
	if err != nil {
		log.Fatalf("错误: 无法打开数据库文件 %s: %v", *dbPath, err)
	}
	defer f.Close()

	db, err = ip2x.New(f)
	if err != nil {
		log.Fatalf("错误: 无法初始化数据库 %s: %v", *dbPath, err)
	}

	if *query != "" {
		runQueryMode(strings.Split(*query, ","))
		return
	}

	http.HandleFunc("/query", queryHandler)
	log.Printf("API 服务启动在 %s，访问 /query?ip=<IP地址>", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatalf("错误: 无法启动服务: %v", err)
	}
}
