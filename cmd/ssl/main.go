package main

import (
	"archive/zip"
	"bytes"
	"context"
	b64 "encoding/base64"
	errors2 "errors"
	"flag"
	"fmt"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Domain struct {
	needApply     bool
	certificateId string
}

type Client struct {
	client *ssl.Client
}

const dvAuthMethod = "DNS_AUTO"

func (c *Client) RestartNginx() (err error) {
	ctxCommand, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(30))
	defer cancel()
	cmd := exec.CommandContext(ctxCommand, "systemctl", "restart", "nginx")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if ctxCommand.Err() == context.DeadlineExceeded {
		log.Println("Command timed out")
	}
	if err != nil {
		log.Println(err)
		return
	}
	if len(stderr.Bytes()) > 0 {
		err = errors2.New(stderr.String())
		log.Println(err)
		return
	}
	return
}

func (c *Client) SaveCertificateToNginx(tmpFile *os.File, saveDir string) (err error) {
	archive, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		log.Println("err:", err)
		return
	}
	defer func(archive *zip.ReadCloser) {
		err = archive.Close()
		if err != nil {
			log.Println("err:", err)
		}
	}(archive)

	for _, f := range archive.File {
		if !strings.HasPrefix(f.Name, "Nginx/") || f.FileInfo().IsDir() {
			continue
		}
		filePath := filepath.Join(saveDir, strings.TrimPrefix(f.Name, "Nginx/"))
		log.Println("unzipping file:", filePath)

		dstFile, e := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if e != nil {
			err = e
			log.Println("err:", err)
			return
		}

		fileInArchive, e := f.Open()
		if e != nil {
			err = e
			log.Println("err:", err)
			return
		}

		if _, err = io.Copy(dstFile, fileInArchive); err != nil {
			log.Println("err:", err)
			return
		}

		dstFile.Close()
		err = fileInArchive.Close()
		if err != nil {
			log.Println("err:", err)
		}
	}
	return
}

func (c *Client) SaveCertificateToTemp(tmpFile *os.File, saveDir string) (err error) {
	archive, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		log.Println("err:", err)
		return
	}
	defer func(archive *zip.ReadCloser) {
		err = archive.Close()
		if err != nil {
			log.Println("err:", err)
		}
	}(archive)

	for _, f := range archive.File {
		filePath := filepath.Join(saveDir, f.Name)
		log.Println("unzipping file:", filePath)

		if f.FileInfo().IsDir() {
			log.Println("creating directory...")
			if err = os.MkdirAll(filePath, os.ModePerm); err != nil {
				log.Println("err:", err)
				return
			}
			continue
		}

		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			log.Println("err:", err)
			return
		}

		dstFile, e := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if e != nil {
			err = e
			log.Println("err:", err)
			return
		}

		fileInArchive, e := f.Open()
		if e != nil {
			err = e
			log.Println("err:", err)
			return
		}

		if _, err = io.Copy(dstFile, fileInArchive); err != nil {
			log.Println("err:", err)
			return
		}

		dstFile.Close()
		err = fileInArchive.Close()
		if err != nil {
			log.Println("err:", err)
		}
	}
	return
}

func (c *Client) DownloadCertificate(certificateId string, saveDir string) (err error) {
	request := ssl.NewDownloadCertificateRequest()
	request.CertificateId = common.StringPtr(certificateId)
	response, err := c.client.DownloadCertificate(request)

	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Println("An API error has returned:", err)
		return
	}
	if err != nil {
		log.Println("err:", err)
		return
	}

	if len(*response.Response.Content) > 0 {
		bs, e := b64.StdEncoding.DecodeString(*response.Response.Content)
		if e != nil {
			err = e
			log.Println("err:", err)
			return
		}
		tmpFile, e := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s.*.zip", certificateId))
		defer os.Remove(tmpFile.Name())
		if e != nil {
			err = e
			log.Println("err:", err)
			return
		}
		_, err = tmpFile.Write(bs)
		if err != nil {
			log.Println("Failed to write to temporary file:", err)
			return
		}

		if saveDir == os.TempDir() {
			_ = c.SaveCertificateToTemp(tmpFile, saveDir)
			return
		}
		_ = c.SaveCertificateToNginx(tmpFile, saveDir)
	}
	return
}

func (c *Client) DescribeCertificate(certificateId string) (deployable bool, err error) {
	request := ssl.NewDescribeCertificateRequest()
	request.CertificateId = common.StringPtr(certificateId)
	response, err := c.client.DescribeCertificate(request)

	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Println("An API error has returned:", err)
		return
	}
	if err != nil {
		log.Println("err:", err)
		return
	}

	deployable = *response.Response.Deployable
	return
}

func (c *Client) DescribeCertificates() (domains map[string]*Domain, err error) {
	request := ssl.NewDescribeCertificatesRequest()
	response, err := c.client.DescribeCertificates(request)

	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Println("An API error has returned:", err)
		return
	}
	if err != nil {
		log.Println("err:", err)
		return
	}

	domains = make(map[string]*Domain)
	for _, v := range response.Response.Certificates {
		certEndTime, _ := time.Parse("2006-01-02 15:04:05", *v.CertEndTime)
		needApply := true
		if certEndTime.Sub(time.Now()) > time.Hour*24*30 {
			needApply = false
		}

		value, ok := domains[*v.Domain]
		if !ok {
			domains[*v.Domain] = &Domain{
				needApply: needApply,
			}
		} else {
			if !needApply && value.needApply {
				domains[*v.Domain] = &Domain{
					needApply: needApply,
				}
			}
		}
	}
	return
}

func (c *Client) ApplyCertificate(domain string) (certificateId string, err error) {
	request := ssl.NewApplyCertificateRequest()
	request.DvAuthMethod = common.StringPtr(dvAuthMethod)
	request.DomainName = common.StringPtr(domain)
	response, err := c.client.ApplyCertificate(request)

	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Println("An API error has returned:", err)
		return
	}
	if err != nil {
		log.Println("err:", err)
		return
	}

	certificateId = *response.Response.CertificateId
	return
}

func NewClient(secretId, secretKey, region string) (client *Client, err error) {
	credential := common.NewCredential(secretId, secretKey)
	sslClient, err := ssl.NewClient(credential, region, profile.NewClientProfile())
	if err != nil {
		log.Println("err:", err)
		return
	}
	client = &Client{
		client: sslClient,
	}
	return
}

func main() {
	secretId := flag.String("secret-id", os.Getenv("TENCENT_SECRET_ID"), "set tencent secret-id")
	secretKey := flag.String("secret-key", os.Getenv("TENCENT_SECRET_KEY"), "set tencent secret-key")
	region := flag.String("region", "", "set region")
	saveDir := flag.String("save-dir", os.TempDir(), "set certificates save path")
	flag.Parse()
	log.Println("save dir:", *saveDir)
	client, err := NewClient(*secretId, *secretKey, *region)
	if err != nil {
		log.Println("err:", err)
		return
	}
	domains, _ := client.DescribeCertificates()
	for domain, v := range domains {
		if !v.needApply {
			continue
		}
		log.Println("need apply:", domain)
		certificateId, err := client.ApplyCertificate(domain)
		if err != nil {
			continue
		}
		i := 0
		for i < 10 {
			i++
			time.Sleep(time.Second)
			deployable, err := client.DescribeCertificate(certificateId)
			if err != nil {
				continue
			}
			if deployable {
				log.Println("apply success:", domain)
				break
			}
		}
		if err = client.DownloadCertificate(certificateId, *saveDir); err != nil {
			log.Println("download certificate success")
		} else {
			log.Println("download certificate error")
		}
	}
	if err = client.RestartNginx(); err != nil {
		log.Println("restart nginx success")
	} else {
		log.Println("restart nginx error")
	}
}
