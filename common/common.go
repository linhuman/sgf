package common

import (
	"encoding/base64"
	"encoding/json"
	"github.com/linhuman/sgf/config"

	//"encoding/hex"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	//"crypto/sha1"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"time"
)

func SliceLen(l interface{}) int {
	if reflect.TypeOf(l).Kind() != reflect.Slice {
		return -1
	}
	ins := reflect.ValueOf(l)
	return ins.Len()
}

func WLog(title string, content interface{}, file string, calldepth int) {
	date := time.Now().Format("2006-01-02")
	dir := config.Entity.LogPath + "/" + time.Now().Format("200601")
	if IsExist(dir) == false {
		err := os.MkdirAll(dir, os.ModePerm)
		if nil != err {
			panic(err)
		}
	}
	var build strings.Builder
	build.WriteString(dir)
	build.WriteString("/")
	file = strings.ReplaceAll(file, "/", "_")
	build.WriteString(file)
	build.WriteString("_")
	build.WriteString(date)
	build.WriteString(".log")
	file = build.String()
	if IsExist(file) {
		fi, err := os.Stat(file)
		if nil == err {
			file_size := fi.Size()
			if file_size > 3072000 { //30mb
				os.Rename(file, file+"_"+time.Now().Format("150405"))
			}
		}
	}
	log_file, err := os.OpenFile(build.String(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	//defer log_file.Close()
	defer file_close(log_file)
	if nil != err {
		panic(err)
	}
	loger := log.New(log_file, "["+title+"]", log.Ldate|log.Ltime|log.Lshortfile)
	loger.Output(calldepth, fmt.Sprintln(content))
}
func WriteLog(title string, content interface{}, file string) {
	WLog(title, content, file, 3)
}
func file_close(log_file *os.File) {
	log_file.Close()
}
func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}
func IsFile(f string) bool {
	fi, e := os.Stat(f)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}

//RSA加密
func RSA_Encrypt(plainText []byte, public_key string) []byte {
	block, _ := pem.Decode([]byte(public_key))
	//x509解码
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	//类型断言
	publicKey := publicKeyInterface.(*rsa.PublicKey)
	//对明文进行加密
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, plainText)
	if err != nil {
		panic(err)
	}
	//返回密文
	return cipherText
}

//RSA解密
func RSA_Decrypt(cipherText []byte, private_key string) []byte {
	block, _ := pem.Decode([]byte(private_key))
	//X509解码
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	//对密文进行解密
	plainText, _ := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherText)
	//返回明文
	return plainText
}
func AES_cbcencrypt(data, pass, iv string) (string, error) {
	block, err := aes.NewCipher([]byte(pass))
	if err != nil {
		return "", err
	}
	src := padding([]byte(data), block.BlockSize())
	rs := make([]byte, len(src))
	blockMode := cipher.NewCBCEncrypter(block, []byte(iv))
	blockMode.CryptBlocks(rs, src)
	data = base64.StdEncoding.EncodeToString(rs)
	return data, nil
}
func AES_cbcdecrype(data, pass, iv string) (string, error) {
	block, err := aes.NewCipher([]byte(pass))
	if err != nil {
		return "", err
	}
	src, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	rs := make([]byte, len(src))

	blockMode := cipher.NewCBCDecrypter(block, []byte(iv))
	blockMode.CryptBlocks(rs, src)
	rs = unpadding(rs)
	return string(rs), nil
}

// 填充数据
func padding(src []byte, blockSize int) []byte {
	padNum := blockSize - len(src)%blockSize
	pad := bytes.Repeat([]byte{byte(padNum)}, padNum)
	return append(src, pad...)
}

// 去掉填充数据
func unpadding(src []byte) []byte {
	n := len(src)
	unPadNum := int(src[n-1])
	return src[:n-unPadNum]
}
func FailOnError(err error, a ...interface{}) {
	if nil != err {
		divide := "------------------------------\n"
		log := err.Error() + "\n" + divide
		for _, v := range a {
			log = log + fmt.Sprintln(v) + divide
		}
		panic("\n" + divide + log)
	}
}
func Fail(a ...interface{}) {
	log := ""
	divide := "------------------------------\n"
	for _, v := range a {
		log = log + fmt.Sprintln(v) + divide
	}
	panic("\n" + divide + log)
}

// 获取正在运行的函数名
func GetFunctionName() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])

	return f.Name()
}
func Strtotime(data_str string) (int64, error) {
	str_arr := strings.Fields(data_str)
	if 0 == len(str_arr) {
		return time.Now().Unix(), nil
	}
	if 1 == len(str_arr) {
		data_str = data_str + " 00:00:00"
	}
	tm, err := time.ParseInLocation("2006-01-02 15:04:05", data_str, time.Local)
	if nil != err {
		return 0, err
	}
	return tm.Unix(), err
}
func Date(format string, timestamp ...int64) string {
	if strings.ContainsAny(format, "Y&m&d&H&i&s") {
		format = strings.Replace(format, "Y", "2006", -1)
		format = strings.Replace(format, "m", "01", -1)
		format = strings.Replace(format, "d", "02", -1)
		format = strings.Replace(format, "H", "15", -1)
		format = strings.Replace(format, "i", "04", -1)
		format = strings.Replace(format, "s", "05", -1)
	}
	if 0 == len(timestamp) {
		return time.Now().Format(format)
	}
	tm := time.Unix(timestamp[0], 0)
	return tm.Format(format)
}
func Integer2str(i interface{}) string {
	switch i.(type) {
	case int:
		return strconv.Itoa(i.(int))
	case int64:
		return strconv.FormatInt(i.(int64), 10)
	case int32:
		return strconv.FormatInt(int64(i.(int32)), 10)
	case int16:
		return strconv.FormatInt(int64(i.(int16)), 10)
	case int8:
		return strconv.FormatInt(int64(i.(int8)), 10)
	case uint:
		return strconv.FormatUint(uint64(i.(uint)), 10)
	case uint8:
		return strconv.FormatUint(uint64(i.(uint8)), 10)
	case uint16:
		return strconv.FormatUint(uint64(i.(uint16)), 10)
	case uint32:
		return strconv.FormatUint(uint64(i.(uint32)), 10)
	case uint64:
		return strconv.FormatUint(i.(uint64), 10)
	default:
		return "0"
	}
}
func Str2Int(s string) (int, error) {
	return strconv.Atoi(s)
}
func Str2Int32(s string) (int32, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	return int32(i), err
}
func Str2Int64(s string) (int64, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	return i, err
}
func Float2str(i interface{}) string {
	switch i.(type) {
	case float32:
		return strconv.FormatFloat(float64(i.(float32)), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(i.(float64), 'f', -1, 32)
	default:
		return "0.00"
	}
}

func ErrRecoverLog(title string, file string) {
	if err := recover(); err != nil {
		//  输出异常信息
		WLog(title, []interface{}{err, string(debug.Stack())}, file, 5)
	}
}
func Echo(a ...interface{}) {
	fmt.Println(a...)
}
func HttpDo(url_string string, method string, params url.Values, headers map[string]string) ([]byte, error) {
	if "" == method {
		method = "GET"
	}
	client := &http.Client{}
	query_str := params.Encode()
	req, err := http.NewRequest(method,
		url_string,
		strings.NewReader(query_str))
	if err != nil {
		return []byte{}, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return []byte{}, err
	}
	body, err := ioutil.ReadAll(resp.Body)

	return body, err
}
func Sleep(seconds time.Duration) {
	time.Sleep(seconds * time.Second)
}

//监听信号
func HandleSignalToStop(f interface{}, args ...interface{}) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM: // 终止进程执行
			signal.Stop(ch)
			if nil == f {
				return
			} else if 1 < len(args) {
				f.(func(...interface{}))(args...)
			} else if 1 == len(args) {
				f.(func(interface{}))(args[0])
			} else {
				f.(func())()
			}
			return
		}
	}
}
func JsonDecode(json_str string, result interface{}) error {
	err := json.Unmarshal([]byte(json_str), result)
	return err
}
func JsonEncode(data interface{}) (string, error) {
	json_byte, err := json.Marshal(data)
	if nil != err {
		return "", err
	}
	return string(json_byte), err
}
