package ipdb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

const IPv4 = 0x01
const IPv6 = 0x02

var (
	ErrFileSize = errors.New("IP Database file size error.")
	ErrMetaData = errors.New("IP Database metadata error.")
	ErrReadFull = errors.New("IP Database ReadFull error.")

	ErrDatabaseError = errors.New("database error")

	ErrIPFormat = errors.New("Query IP Format error.")

	ErrNoSupportLanguage = errors.New("language not support")
	ErrNoSupportIPv4     = errors.New("IPv4 not support")
	ErrNoSupportIPv6     = errors.New("IPv6 not support")

	ErrDataNotExists = errors.New("data is not exists")

	ErrInvalidIP = errors.New("invalid ip address")
)

type MetaData struct {
	Build     int64          `json:"build"`
	IPVersion uint16         `json:"ip_version"`
	Languages map[string]int `json:"languages"`
	NodeCount int            `json:"node_count"`
	TotalSize int            `json:"total_size"`
	Fields    []string       `json:"fields"`
}

type reader struct {
	fileSize  int
	nodeCount int
	v4offset  int

	meta MetaData
	data []byte

	refType map[string]string
}

// IP range: ip_start ip_end
type ipNetwork struct {
	ipEnd string
	ipNet *net.IPNet
}

func newReader(name string, obj interface{}) (*reader, error) {
	var err error
	var fileInfo os.FileInfo
	fileInfo, err = os.Stat(name)
	if err != nil {
		return nil, err
	}
	fileSize := int(fileInfo.Size())
	if fileSize < 4 {
		return nil, ErrFileSize
	}
	body, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, ErrReadFull
	}
	var meta MetaData
	metaLength := int(binary.BigEndian.Uint32(body[0:4]))
	if fileSize < (4 + metaLength) {
		return nil, ErrFileSize
	}
	if err := json.Unmarshal(body[4:4+metaLength], &meta); err != nil {
		return nil, err
	}
	if len(meta.Languages) == 0 || len(meta.Fields) == 0 {
		return nil, ErrMetaData
	}
	if fileSize != (4 + metaLength + meta.TotalSize) {
		return nil, ErrFileSize
	}

	var dm map[string]string
	if obj != nil {
		t := reflect.TypeOf(obj).Elem()
		dm = make(map[string]string, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			k := t.Field(i).Tag.Get("json")
			dm[k] = t.Field(i).Name
		}
	}

	db := &reader{
		fileSize:  fileSize,
		nodeCount: meta.NodeCount,

		meta:    meta,
		refType: dm,

		data: body[4+metaLength:],
	}

	if db.v4offset == 0 {
		node := 0
		for i := 0; i < 96 && node < db.nodeCount; i++ {
			if i >= 80 {
				node = db.readNode(node, 1)
			} else {
				node = db.readNode(node, 0)
			}
		}
		db.v4offset = node
	}

	return db, nil
}

func (db *reader) Find(addr, language string) ([]string, error) {
	return db.find1(addr, language)
}

func (db *reader) FindMap(addr, language string) (map[string]string, error) {

	data, err := db.find1(addr, language)
	if err != nil {
		return nil, err
	}

	// Add node and ip range
	info := map[string]string{
		"node":     data[0],
		"ip_start": data[1],
		"ip_end":   data[2],
	}
	for k, v := range data[3:] {
		info[db.meta.Fields[k]] = v
	}

	return info, nil
}

func (db *reader) find0(addr string) ([]byte, error) {

	var (
		err     error
		node    int
		ipRange *ipNetwork
	)
	ipv := net.ParseIP(addr)
	if ip := ipv.To4(); ip != nil {
		if !db.IsIPv4Support() {
			return nil, ErrNoSupportIPv4
		}

		node, ipRange, err = db.search(ip, 32)
	} else if ip := ipv.To16(); ip != nil {
		if !db.IsIPv6Support() {
			return nil, ErrNoSupportIPv6
		}

		node, ipRange, err = db.search(ip, 128)
	} else {
		return nil, ErrIPFormat
	}

	if err != nil || node < 0 {
		return nil, err
	}

	body, err := db.resolve(node)
	if err != nil {
		return nil, err
	}

	// Add node and ip range
	var buf bytes.Buffer
	buf.WriteString(strconv.Itoa(node))
	buf.WriteString("\t")
	buf.WriteString(ipRange.ipNet.IP.String())
	buf.WriteString("\t")
	buf.WriteString(ipRange.ipEnd)
	buf.WriteString("\t")
	buf.Write(body)
	body = buf.Bytes()

	return body, nil
}

func (db *reader) find1(addr, language string) ([]string, error) {

	off, ok := db.meta.Languages[language]
	if !ok {
		return nil, ErrNoSupportLanguage
	}

	body, err := db.find0(addr)
	if err != nil {
		return nil, err
	}

	// tmp[0:3] is node, ip_start, ip_end
	off += 3

	str := (*string)(unsafe.Pointer(&body))
	tmp := strings.Split(*str, "\t")

	if (off + len(db.meta.Fields)) > len(tmp) {
		return nil, ErrDatabaseError
	}

	// ip_start ip_end + ipip body
	tmp = append(tmp[0:3], tmp[off:off+len(db.meta.Fields)]...)

	return tmp, nil
}

func (db *reader) search(ip net.IP, bitCount int) (int, *ipNetwork, error) {

	var (
		node    int
		addr    int
		mask    int
		ipRange = &ipNetwork{}
	)

	if bitCount == 32 {
		addr = int(ip[0])<<24 | int(ip[1])<<16 | int(ip[2])<<8 | int(ip[3])
		node = db.v4offset
	} else {
		node = 0
	}

	for i := 0; i < bitCount; i++ {
		if node > db.nodeCount {
			if bitCount == 32 {
				// ip_end
				addr = addr>>(bitCount-i)<<(bitCount-i) + 1<<(bitCount-i) - 1
				mask = i
			}
			break
		}

		node = db.readNode(node, ((0xFF&int(ip[i>>3]))>>uint(7-(i%8)))&1)
	}

	if bitCount == 32 {
		// Only support IPv4
		ipRange.ipEnd = db.IPv4String(addr)
		_, ipRange.ipNet, _ = net.ParseCIDR(fmt.Sprintf("%s/%d", ipRange.ipEnd, mask))
	}

	if node > db.nodeCount {
		return node, ipRange, nil
	}

	return -1, ipRange, ErrDataNotExists
}

func (db *reader) readNode(node, index int) int {
	off := node*8 + index*4
	return int(binary.BigEndian.Uint32(db.data[off : off+4]))
}

func (db *reader) resolve(node int) ([]byte, error) {
	resolved := node - db.nodeCount + db.nodeCount*8
	if resolved >= db.fileSize {
		return nil, ErrDatabaseError
	}

	size := int(binary.BigEndian.Uint16(db.data[resolved : resolved+2]))
	if (resolved + 2 + size) > len(db.data) {
		return nil, ErrDatabaseError
	}
	b := db.data[resolved+2 : resolved+2+size]

	return b, nil
}

func (db *reader) IsIPv4Support() bool {
	return (int(db.meta.IPVersion) & IPv4) == IPv4
}

func (db *reader) IsIPv6Support() bool {
	return (int(db.meta.IPVersion) & IPv6) == IPv6
}

func (db *reader) Build() time.Time {
	return time.Unix(db.meta.Build, 0).In(time.UTC)
}

func (db *reader) Languages() []string {
	ls := make([]string, 0, len(db.meta.Languages))
	for k := range db.meta.Languages {
		ls = append(ls, k)
	}
	return ls
}

// Convert int to ipv4.string
func (db *reader) IPv4String(n int) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(n>>24)&0xFF, byte(n>>16)&0xFF, byte(n>>8)&0xFF, byte(n)&0xFF)
}

// Convert ipv4.string to int
func (db *reader) IPv4Int(s string) (int, error) {
	if ip := net.ParseIP(s).To4(); ip != nil {
		return int(ip[0])<<24 | int(ip[1])<<16 | int(ip[2])<<8 | int(ip[3]), nil
	}
	return 0, ErrInvalidIP
}

// Export ipv4.ipdb to txtx
func (db *reader) IPv4TXTX(lang string) error {
	var (
		err    error
		data   []string
		ipInfo []string
	)

	for ip := 0; ip < 1<<32; {
		if data, err = db.find1(db.IPv4String(ip), lang); err != nil {
			return err
		}
		if ip, err = db.IPv4Int(data[2]); err != nil {
			return err
		}

		ip += 1

		// 1.1.9.0	1.1.63.255	China	Guangdong	*	*	ChinaTelecom
		for i := 0; i < len(data); i++ {
			if data[i] == "" {
				data[i] = "*"
			}
		}

		// first
		if ipInfo == nil {
			ipInfo = data
			continue
		}

		// merge same node
		if ipInfo[0] == data[0] {
			ipInfo[2] = data[2]
			continue
		}

		fmt.Println(strings.Join(ipInfo[1:], "\t"))
		ipInfo = data
	}

	if ipInfo != nil {
		fmt.Println(strings.Join(ipInfo[1:], "\t"))
	}

	return nil
}
