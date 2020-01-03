package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"os"
	"strconv"
)

const (
	primeNumFile = "prime.file"
	primeMax     = 100004
)

// PrimeNum : 与查找素数相关的类
type PrimeNum struct {
	primeGroup []int64
}

func (p *PrimeNum) load() int {
	file, err1 := os.OpenFile(primeNumFile, os.O_RDONLY, 0666)
	defer func() { file.Close() }()

	if err1 != nil {
		return 1
	}

	reader := bufio.NewReader(file)

	for {
		if bLine, _, err := reader.ReadLine(); err == nil {
			v, _ := strconv.Atoi(string(bLine))
			p.primeGroup = append(p.primeGroup, int64(v))
		} else {
			if err != io.EOF {
				return 2
			}
			break
		}
	}

	return 0
}

func (p *PrimeNum) write() int {
	file, err := os.OpenFile(primeNumFile, os.O_CREATE|os.O_WRONLY, 0666)
	defer func() { file.Close() }()

	if err != nil {
		return 3
	}

	for _, v := range p.primeGroup {
		sz, err := file.Write([]byte(strconv.FormatInt(v, 10) + "\n"))
		if err != nil && sz == 0 {
			return 4
		}
	}
	return 0
}

func (p *PrimeNum) generate() {

	var flag []bool = make([]bool, primeMax)
	for i := 3; i < primeMax; i += 2 {
		for j := 3; i*j < primeMax; j += 2 {
			flag[i*j] = true
		}
	}

	p.primeGroup = []int64{2}
	for index, value := range flag {
		if index >= 2 && (index&1 > 0) && value == false {
			p.primeGroup = append(p.primeGroup, int64(index))
		}
	}

}

// JudgePrime 判断一个小于10^12的数是否为素数
func (p *PrimeNum) JudgePrime(num int64) bool {
	if (num != 2 && num%2 == 0) || num == 1 {
		return false
	}
	for _, v := range p.primeGroup {
		if v >= num {
			break
		}
		if num%v == 0 {
			return false
		}
	}
	return true

}

//计算 a^n mod (mod)
func powMod(a int64, n int64, mod int64) int64 {
	if n == 0 {
		return 1
	}
	bs := powMod(a, n/2, mod)
	res := big.NewInt(bs)
	res = res.Mul(res, res)
	if n&1 > 0 {
		res = res.Mul(res, big.NewInt(a))
	}
	return res.Mod(res, big.NewInt(mod)).Int64()
}

// 计算ax+by == gcd(a,b)，假定a,b互素
func exgcd(a int64, b int64) (int64, int64) {
	if b == 0 {
		return 1, 0
	}
	x, y := exgcd(b, a%b)
	return y, (x - (a/b)*y)
}

// 计算a模mod的逆元，mod为质素
func inv(a int64, mod int64) int64 {
	x, _ := exgcd(a, mod)
	return (x + mod) % mod
}

//计算 a^x=b (mod (mod)) 返回 x
// x = im+j, m = ceil(sqrt(mod));
// i,j belongs [0,m)
// 枚举 b*inv(a)^j
// 判断 (a^m)^i
func bsgs(a int64, b int64, mod int64) int64 {
	m := int64(math.Ceil(math.Sqrt(float64(mod))))
	// 记录所有b*inv(a)^j -> j
	hashtable := make(map[int64]int64)
	invA := inv(a, mod)
	t1 := big.NewInt(b % mod)
	for j := int64(0); j < m; j++ {
		hashtable[t1.Int64()] = int64(j)
		t1 = t1.Mul(t1, big.NewInt(invA)).Mod(t1, big.NewInt(mod))
	}
	t2 := big.NewInt(1)
	apm := powMod(a, m, mod)
	for i := int64(0); i < m; i++ {
		j, exist := hashtable[t2.Int64()]
		if exist {
			return int64(int64(i)*m + j)
		}
		t2.Mul(t2, big.NewInt(apm)).Mod(t2, big.NewInt(mod))
	}

	return -1
}

func debug() {
	p := new(PrimeNum)
	ret := p.load()
	if ret != 0 {
		p.generate()
		p.write()
	}
	fmt.Println(p.JudgePrime(100003))
	fmt.Println(p.JudgePrime(1000000007))
	fmt.Println(p.JudgePrime(1000000011))

	//r := powMod(10000000000000000, 2, 1000000007)
	//ia := big.NewInt(10000000000000000)
	//ib := big.NewInt(10000000000000000)
	//x := ia.Mul(ia, ib).Sub(ia, big.NewInt(r)).Mod(ia, big.NewInt(1000000007)).String()
	//fmt.Println(r, x)

	//	t := inv(10000000000000000, 1000003)
	//	ia := big.NewInt(10000000000000000)
	//	r := ia.Mul(ia, big.NewInt(t)).Mod(ia, big.NewInt(1000003)).Int64()
	//	fmt.Println(t, r)

	//	r := powMod(10000000000000000, 1564, 1000000007)
	//	x := bsgs(10000000000000000, 773909255, 1000000007)
	//	fmt.Println(x)
}

type httpAccess struct {
	pr PrimeNum
}

func (access *httpAccess) HandleJudgePrime(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	strNum := vars.Get("num")
	intNum, err := strconv.ParseInt(strNum, 10, 64)
	flag := 0
	if err != nil {
		flag = 1
	}
	if intNum >= 1000000000000 || intNum < 1 {
		flag = 2
	}
	if flag == 0 {
		boolPrime := access.pr.JudgePrime(intNum)
		if boolPrime != true {
			flag = 3
		}
	}
	msg := []string{"输入正确", "格式错误", "输入数字范围错误", "不是素数"}
	resp := struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
	}{flag, msg[flag]}

	j, _ := json.Marshal(resp)
	w.Write(j)
}

func (access *httpAccess) HandlePowMod(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	strA := vars.Get("a")
	strN := vars.Get("n")
	strM := vars.Get("m")
	intA, err1 := strconv.ParseInt(strA, 10, 64)
	intN, err2 := strconv.ParseInt(strN, 10, 64)
	intM, err3 := strconv.ParseInt(strM, 10, 64)
	flag := 0
	if err1 != nil || err2 != nil || err3 != nil {
		flag = 1
	}
	judgeRange := func(a int64) bool {
		return a > 0 && a < 1000000000000
	}
	if !(judgeRange(intA) && judgeRange(intN) && judgeRange(intM)) {
		flag = 2
	}

	if !access.pr.JudgePrime(intM) {
		flag = 3
	}

	pm := int64(0)

	if flag == 0 {
		pm = powMod(intA, intN, intM)
	} else {
		pm = int64(-1)
	}

	msg := []string{"输入正确", "格式错误", "输入数字范围错误", "m不是素数"}
	resp := struct {
		Status int    `json:"status"`
		Result string `json:"result"`
		Msg    string `json:"msg"`
	}{flag, strconv.FormatInt(pm, 10), msg[flag]}

	j, _ := json.Marshal(resp)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(j)
}

func (access *httpAccess) HandleBsgs(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	strA := vars.Get("a")
	strB := vars.Get("b")
	strM := vars.Get("m")
	intA, err1 := strconv.ParseInt(strA, 10, 64)
	intB, err2 := strconv.ParseInt(strB, 10, 64)
	intM, err3 := strconv.ParseInt(strM, 10, 64)
	flag := 0
	if err1 != nil || err2 != nil || err3 != nil {
		flag = 1
	}
	judgeRange := func(a int64) bool {
		return a > 0 && a < 1000000000000
	}
	if !(judgeRange(intA) && judgeRange(intB) && judgeRange(intM)) {
		flag = 2
	}

	if !access.pr.JudgePrime(intM) {
		flag = 3
	}

	pm := int64(0)
	strResult := ""
	if flag == 0 {
		pm = bsgs(intA, intB, intM)
		if pm == -1 {
			strResult = "不存在"
		} else {
			strResult = strconv.FormatInt(pm, 10)
		}
	}

	msg := []string{"输入正确", "格式错误", "输入数字范围错误", "m不是素数"}
	resp := struct {
		Status int    `json:"status"`
		Result string `json:"result"`
		Msg    string `json:"msg"`
	}{flag, strResult, msg[flag]}

	j, _ := json.Marshal(resp)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(j)
}

func main() {
	access := new(httpAccess)
	access.pr.load()

	http.HandleFunc("/prime", access.HandleJudgePrime)
	http.HandleFunc("/powmod", access.HandlePowMod)
	http.HandleFunc("/bsgs", access.HandleBsgs)

	go http.ListenAndServe("127.0.0.1:40000", nil)
	http.ListenAndServe("127.0.0.1:8080", http.FileServer(http.Dir("./")))

}
