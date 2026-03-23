package ip

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// GeoLite2 免费版下载地址
	geolite2CityURL = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=%s&suffix=tar.gz"
	geolite2ISPURL  = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key=%s&suffix=tar.gz"

	// 默认下载超时
	downloadTimeout = 10 * time.Minute

	// 更新检查间隔（1天）
	updateInterval = 24 * time.Hour
)

// GeoLite2Downloader GeoLite2 数据库下载器
type GeoLite2Downloader struct {
	licenseKey string
	dataDir    string
	client     *http.Client
	stopCh     chan struct{}
}

// GeoLite2Config GeoLite2 下载配置
type GeoLite2Config struct {
	LicenseKey string // MaxMind License Key
	DataDir    string // 数据文件存放目录
}

// NewGeoLite2Downloader 创建下载器
func NewGeoLite2Downloader(cfg GeoLite2Config) *GeoLite2Downloader {
	if cfg.DataDir == "" {
		cfg.DataDir = "./data"
	}

	return &GeoLite2Downloader{
		licenseKey: cfg.LicenseKey,
		dataDir:    cfg.DataDir,
		client: &http.Client{
			Timeout: downloadTimeout,
		},
		stopCh: make(chan struct{}),
	}
}

// Download 下载并解压 GeoLite2 数据库
func (d *GeoLite2Downloader) Download() error {
	if d.licenseKey == "" {
		return fmt.Errorf("license key is required")
	}

	// 确保目录存在
	if err := os.MkdirAll(d.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// 下载 City 数据库
	cityDBPath, err := d.downloadAndExtract("City", geolite2CityURL)
	if err != nil {
		logx.Errorf("Failed to download City database: %v", err)
		// 继续尝试下载 ISP
	} else {
		logx.Infof("City database downloaded: %s", cityDBPath)
	}

	// 下载 ISP (ASN) 数据库
	ispDBPath, err := d.downloadAndExtract("ASN", geolite2ISPURL)
	if err != nil {
		logx.Errorf("Failed to download ISP database: %v", err)
	} else {
		logx.Infof("ISP database downloaded: %s", ispDBPath)
	}

	return nil
}

// downloadAndExtract 下载并解压数据库
func (d *GeoLite2Downloader) downloadAndExtract(dbType, urlTemplate string) (string, error) {
	url := fmt.Sprintf(urlTemplate, d.licenseKey)

	logx.Infof("Downloading GeoLite2 %s database...", dbType)

	// 下载文件
	resp, err := d.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned status: %d", resp.StatusCode)
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("geolite2-%s-*.tar.gz", dbType))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入临时文件
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// 解压文件
	dbPath, err := d.extract(tmpFile.Name(), dbType)
	if err != nil {
		return "", fmt.Errorf("extraction failed: %w", err)
	}

	return dbPath, nil
}

// extract 解压 tar.gz 文件
func (d *GeoLite2Downloader) extract(tarGzPath, dbType string) (string, error) {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var dbPath string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// 查找 .mmdb 文件
		if strings.HasSuffix(header.Name, ".mmdb") {
			// 构建目标路径
			filename := filepath.Base(header.Name)
			if dbType == "City" {
				filename = "GeoLite2-City.mmdb"
			} else if dbType == "ASN" {
				filename = "GeoLite2-ASN.mmdb"
			}
			targetPath := filepath.Join(d.dataDir, filename)

			// 创建目标文件
			targetFile, err := os.Create(targetPath)
			if err != nil {
				return "", err
			}

			// 复制内容
			if _, err := io.Copy(targetFile, tarReader); err != nil {
				targetFile.Close()
				return "", err
			}
			targetFile.Close()

			dbPath = targetPath
			break
		}
	}

	if dbPath == "" {
		return "", fmt.Errorf("mmdb file not found in archive")
	}

	return dbPath, nil
}

// StartAutoUpdate 启动自动更新
func (d *GeoLite2Downloader) StartAutoUpdate() {
	go func() {
		ticker := time.NewTicker(updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				logx.Info("Checking for GeoLite2 database updates...")
				if err := d.Download(); err != nil {
					logx.Errorf("Auto update failed: %v", err)
				}
			case <-d.stopCh:
				logx.Info("GeoLite2 auto update stopped")
				return
			}
		}
	}()

	logx.Infof("GeoLite2 auto update started, interval: %v", updateInterval)
}

// StopAutoUpdate 停止自动更新
func (d *GeoLite2Downloader) StopAutoUpdate() {
	close(d.stopCh)
}

// GetDBPaths 获取数据库文件路径
func (d *GeoLite2Downloader) GetDBPaths() (cityDB, ispDB string) {
	cityDB = filepath.Join(d.dataDir, "GeoLite2-City.mmdb")
	ispDB = filepath.Join(d.dataDir, "GeoLite2-ASN.mmdb")

	// 检查文件是否存在
	if _, err := os.Stat(cityDB); err != nil {
		cityDB = ""
	}
	if _, err := os.Stat(ispDB); err != nil {
		ispDB = ""
	}

	return
}

// IsAvailable 检查数据库是否可用
func (d *GeoLite2Downloader) IsAvailable() bool {
	cityDB, _ := d.GetDBPaths()
	return cityDB != ""
}
