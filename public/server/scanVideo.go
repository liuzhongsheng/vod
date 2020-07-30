package main
import (
	"os"
	"fmt"
	"path/filepath"
	"path"
	"crypto/sha256"
    "io"
	"strings"
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql"
	"crypto/md5"
	"encoding/hex"
	"time"
)
// 定义配置模型
type VodConfig struct{
	Name string
	Group string
	Value string
}

// 定义队列模型
type VodQueue struct{
	Createtime   int64
	Updatetime   int64
    Md5          string
    Name         string
    Id           int     `gorm:"AUTO_INCREMENT"` // 自增
    LocalPath    string
    Suffix  	 string
    Hash 		 string
}
//返回一个32位md5加密后的字符串
func GetMD5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

//返回一个16位md5加密后的字符串
func Get16MD5Encode(data string) string{
	return GetMD5Encode(data)[8:24]
}

// 获取文件hash值
func gethash(path string) (hash string) {
    file, err := os.Open(path)
    if err == nil {
        h_ob := sha256.New()
        _, err := io.Copy(h_ob, file)
        if err == nil {
            hash := h_ob.Sum(nil)
            hashvalue := hex.EncodeToString(hash)
            return hashvalue
        } else {
            return "something wrong when use sha256 interface..."
        }
    } else {
        fmt.Printf("failed to open %s\n", path)
    }
    defer file.Close()
    return
}
func main(){
	// 连接数据库 
	mysql_conn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", "vod","Z3pHbFWbyWjiGsmC", "120.77.152.67", 3306, "vod")
	fmt.Println("数据库连接:",mysql_conn)
	db, err := gorm.Open("mysql", mysql_conn)
	db.SingularTable(true)
	// 启用Logger，显示详细日志
	// db.LogMode(true)
  	defer db.Close()
  	if err != nil {
        panic(err)
    }
    fmt.Println("初始化数据库成功......")
    fileDirPath := VodConfig{}
    // 获取目录路径
 	db.Find(&fileDirPath,"`group` = ? and `name` = ?", "video_synchro","local_path")
    fmt.Println("当前目录为：",fileDirPath.Value)

    // 获取支持的格式
    configSuffix := VodConfig{}
    // 获取目录路径
 	db.Find(&configSuffix,"`group` = ? and `name` = ?", "video_synchro","video_suffix")
    fmt.Println("当前允许的格式为：",configSuffix.Value)

	//获取当前目录下的所有文件或目录信息
	filepath.Walk(fileDirPath.Value,func(runPath string, info os.FileInfo, err error) error{
		
		f, _ := os.Stat(runPath)
		// 检测当前是否为文件
		if(!f.IsDir()){
			fmt.Println(runPath) //打印path信息
			fmt.Println(info.Name()) //打印文件或目录名

			// 获取文件后缀，并转为小写
			filenameWithSuffix := path.Base(runPath)
			fileSuffix := strings.ToLower(path.Ext(filenameWithSuffix))
			if(strings.Contains(configSuffix.Value, fileSuffix)){
				var filenameOnly string
			    filenameOnly = strings.TrimSuffix(info.Name(), fileSuffix)
			    tempData := VodQueue{}
			    hash := gethash(runPath);
				fmt.Println("文件名：",filenameOnly,"文件格式",fileSuffix,"文件密文",Get16MD5Encode(filenameOnly),"hash值",hash)
				db.Find(&tempData,"`hash` = ?",hash);
				fmt.Println(tempData)
				if(tempData.Name != ""){
					fmt.Println("文件存在")
				}else{
					fmt.Println("文件不存在写入")
					data := VodQueue{Name: filenameOnly, Md5: Get16MD5Encode(filenameOnly), Createtime: time.Now().Unix(),LocalPath:runPath,Suffix:fileSuffix,Hash:hash}
					db.Create(&data)
					fmt.Println(data)
				}
			}
		}
		return nil
	})
} 