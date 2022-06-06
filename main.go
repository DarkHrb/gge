package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type User struct {
	Name    string
	Age     int
	Sex     string
	Address string
}

var db *gorm.DB

func main() {
	//设置全局的logger,这个logger在执行每个sql语句的时候会打印每一行的sql
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,        // 禁用彩色打印
		},
	)

	dsn := "host=localhost user=postgres password=dark415. dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		}})
	if err != nil {
		log.Println(err)
	}

	// 迁移 schema --- 建表
	//err = db.AutoMigrate(&User{})
	//if err != nil {
	//	fmt.Println("err:", err)
	//}

	r := gin.Default()

	r.GET("/getData/all", func(c *gin.Context) {
		var users []User

		// 获取全部记录
		db.Find(&users)

		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": users,
		})
	})

	r.GET("/getData/page", func(c *gin.Context) {
		current := c.Query("current")
		size := c.Query("size")

		var users []User

		currentPage, _ := strconv.Atoi(current)
		pageSize, _ := strconv.Atoi(size)

		// 分页查询
		db.Limit(pageSize).Offset((currentPage - 1) * pageSize).Order("age").Find(&users)
		//err := db.Limit(pageSize).Offset((currentPage - 1) * pageSize).Find(&list).Error // 查询分页后所需的数据

		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": users,
		})
	})

	r.POST("/upload", func(c *gin.Context) {
		// Single file
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "上传文件出错",
			})
			return
		}

		reader, err := excelize.OpenReader(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "读取excel文件失败",
			})
			return
		}

		//解析excel的数据
		users, lxRrr := readExcel(reader)
		if lxRrr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "解析excel文件失败",
			})
			return
		}

		db.CreateInBatches(users, len(users))

		// Upload the file to specific dst.
		//c.SaveUploadedFile(file, dst)

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "上传成功",
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

//ReadExcel .读取excel 转成切片
func readExcel(xlsx *excelize.File) ([]User, error) {
	//根据名字获取cells的内容，返回的是一个[][]string
	rows, err := xlsx.GetRows(xlsx.GetSheetName(xlsx.GetActiveSheetIndex()))
	if err != nil {
		return nil, err
	}
	//声明一个数组
	var users []User
	for i, row := range rows {
		// 去掉第一行是excel表头部分
		if i == 0 {
			continue
		}
		var data User
		for k, v := range row {
			// 第一列是name
			if k == 0 {
				data.Name = v
			} else if k == 1 { // 第二列是age
				data.Age, _ = strconv.Atoi(v)
			} else if k == 2 { // 第三列是sex
				data.Sex = v
			} else if k == 3 {
				data.Address = v
			}
		}
		//将数据追加到集合中
		users = append(users, data)
	}

	return users, nil
}
