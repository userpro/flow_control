/*
 * @Author: dongzhzheng
 * @Date: 2021-03-29 16:45:44
 * @LastEditTime: 2021-03-30 10:14:32
 * @LastEditors: dongzhzheng
 * @FilePath: /flow_control/flow_control.go
 * @Description:
 */

package flowcontrol

import "crypto/md5"

var (
	defaultFlowControlOptions = FlowControllerOptions{
		Radio: 0,
		Hash:  defaultTafHash,
	}

	md5Ctx = md5.New()
)

// defaultTafHash
func defaultTafHash(key string) uint32 {
	md5Ctx.Reset()
	_, _ = md5Ctx.Write([]byte(key))
	cipherStr := md5Ctx.Sum(nil)
	hash := (int32(cipherStr[3]&0xFF) << 24) |
		(int32(cipherStr[2]&0xFF) << 16) |
		(int32(cipherStr[1]&0xFF) << 8) |
		(int32(cipherStr[0] & 0xFF))
	return uint32(hash)
}

// FlowControllerOption 选项函数
type FlowControllerOption func(*FlowControllerOptions)

// FlowControllerOptions 可配置项
type FlowControllerOptions struct {
	Radio uint32
	Hash  func(string) uint32
}

// WithForwardRadio 划分比例
func WithForwardRadio(r uint32) FlowControllerOption {
	return func(fopt *FlowControllerOptions) {
		fopt.Radio = r
	}
}

// WithHashFunc hash函数
func WithHashFunc(h func(string) uint32) FlowControllerOption {
	return func(fopt *FlowControllerOptions) {
		fopt.Hash = h
	}
}

// FlowController 流量控制
type FlowController struct {
	Radio uint32
	Hash  func(string) uint32
}

// Forward 是否转发
func (f FlowController) Forward(key string) bool {
	return (f.Hash(key) & 0x7fffffff % 100) < f.Radio
}

// New ...
func New(opts ...FlowControllerOption) *FlowController {
	options := defaultFlowControlOptions
	for _, o := range opts {
		o(&options)
	}

	return &FlowController{
		Radio: options.Radio,
		Hash:  options.Hash,
	}
}
