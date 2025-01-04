package bolt

import (
	"encoding/json"
	"errors"
	"time"

	"speedtest/database/schema"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	"go.etcd.io/bbolt"
)

const (
	// 数据存储的桶名称
	dataBucketName = `speedtest-ex`
)

type Storage struct {
	db *bbolt.DB
}

var (
	logDebug = logger.Logw
	logInfo  = logger.LogInfo
	logWarn  = logger.LogWarning
	logErr   = logger.LogError
)

// OpenDatabase 打开一个 BoltDB 数据库
func OpenDatabase(dbFilePath string) *Storage {
	db, err := bbolt.Open(dbFilePath, 0666, nil)
	if err != nil {
		logErr("Failed to open BoltDB file: %s", err)
		panic(err) // 直接终止程序，确保问题被及时发现
	}
	return &Storage{db: db}
}

// SaveTelemetry 插入一条 TelemetryData 数据记录
func (s *Storage) SaveTelemetry(data *schema.TelemetryData) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		// 设置时间戳
		data.Timestamp = time.Now()

		// 序列化为 JSON 格式
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return err
		}

		// 创建或获取存储桶
		bucket, err := tx.CreateBucketIfNotExists([]byte(dataBucketName))
		if err != nil {
			return err
		}

		// 根据 UUID 存储数据
		return bucket.Put([]byte(data.UUID), dataBytes)
	})
}

// GetTelemetryByUUID 根据 UUID 获取单条 TelemetryData 数据
func (s *Storage) GetTelemetryByUUID(id string) (*schema.TelemetryData, error) {
	var telemetry schema.TelemetryData

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dataBucketName))
		if bucket == nil {
			return errors.New("storage bucket does not exist")
		}

		// 获取数据
		dataBytes := bucket.Get([]byte(id))
		if dataBytes == nil {
			return errors.New("record not found")
		}

		// 反序列化 JSON 数据
		return json.Unmarshal(dataBytes, &telemetry)
	})

	return &telemetry, err
}

// GetLastNRecords 获取最新的 N 条 TelemetryData 数据
func (s *Storage) GetLastNRecords(limit int) ([]schema.TelemetryData, error) {
	var records []schema.TelemetryData

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dataBucketName))
		if bucket == nil {
			return errors.New("storage bucket does not exist")
		}

		cursor := bucket.Cursor()
		_, dataBytes := cursor.Last()

		for len(records) < limit {
			if dataBytes == nil {
				break
			}

			var record schema.TelemetryData
			if err := json.Unmarshal(dataBytes, &record); err != nil {
				return err
			}

			records = append(records, record)
			_, dataBytes = cursor.Prev()
		}

		return nil
	})

	logInfo("Fetched %d records from storage", len(records))
	return records, err
}

// GetAllTelemetry 获取所有 TelemetryData 数据 (仅用于调试)
func (s *Storage) GetAllTelemetry() ([]schema.TelemetryData, error) {
	var records []schema.TelemetryData

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dataBucketName))
		if bucket == nil {
			return errors.New("storage bucket does not exist")
		}

		cursor := bucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			var record schema.TelemetryData
			if err := json.Unmarshal(value, &record); err != nil {
				return err
			}
			records = append(records, record)
		}

		return nil
	})

	// 输出调试日志
	for _, record := range records {
		logDebug("Record: UUID: %s, Timestamp: %s, IPAddress: %s, ISPInfo: %s, Extra: %s, UserAgent: %s, Language: %s, Download: %s, Upload: %s, Ping: %s, Jitter: %s, Log: %s",
			record.UUID,
			record.Timestamp.Format(time.RFC3339),
			record.IPAddress,
			record.ISPInfo,
			record.Extra,
			record.UserAgent,
			record.Language,
			record.Download,
			record.Upload,
			record.Ping,
			record.Jitter,
			record.Log,
		)
	}

	return records, err
}
