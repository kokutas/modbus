package modbus

import (
	"context"
	"fmt"
	"time"
)

// 功能码常量 bit access
const (
	READ_COILS           byte = 1 // 1(0x01)
	READ_DISCRETE_INPUTS byte = 2 // 2(0x02)
	WRITE_SINGLE_COIL    byte = 5 // 5(0x05)
)

// 功能码常量 16-bit access
const (
	READ_HOLDING_REGISTERS byte = 3 // 3(0x03)
	READ_INPUT_REGISTERS   byte = 4 // 4(0x04)
	WRITE_SINGLE_REGISTER  byte = 6 // 6(0x06)
)
const (
	READ_EXCEPTION_STATUS byte = 7 // 7(0x07)
)

// modbus function
//
//go:generate mockery -name Slaver
type Slaver interface {
	// 读取远程设备中线圈的1-2000的状态(开关量OFF/ON)
	// 功能码: 1字节,01 (0x01)
	// address: 2字节,线圈起始地址,寻址范围[0x0000-0xFFFF]
	// quantity: 2字节,线圈数量[0x0001-0x07D0]
	ReadCoils(ctx context.Context, address, quantity uint16) (results []byte, err error)
	// 读取远程设备中离散输入的1-2000的状态(只读)
	// 功能码: 1字节,02 (0x02)
	// address: 2字节,线圈起始地址,寻址范围[0x0000-0xFFFF]
	// quantity: 2字节,线圈数量[0x0001-0x07D0]
	ReadDiscreteInputs(ctx context.Context, address, quantity uint16) (results []byte, err error)
	// 读取远程设备中1-125保持寄存器内容
	// 功能码: 1字节,03 (0x03)
	// address: 2字节,寄存器起始地址,寻址范围[0x0000-0xFFFF]
	// quantity: 2字节,寄存器数量[0x0001-0x007D]
	ReadHoldingRegisters(ctx context.Context, address, quantity uint16) (results []byte, err error)
	// 读取远程设备中1-125输入寄存器内容(只读)
	// 功能码: 1字节,04 (0x04)
	// address: 2字节,寄存器起始地址,寻址范围[0x0000-0xFFFF]
	// quantity: 2字节,寄存器数量[0x0001-0x007D]
	ReadInputRegisters(ctx context.Context, address, quantity uint16) (results []byte, err error)
	// 在远程设备中写单个线圈(开关量)
	// 功能码: 1字节,05 (0x05)
	// address: 2字节,寄存器起始地址,寻址范围[0x0000-0xFFFF]
	// value: 2字节,数据(开关量)[OFF/ON (0x0000/0xFF00)]
	WriteSingleCoil(ctx context.Context, address, value uint16) (results []byte, err error)
	// 在远程设备中写单个保持寄存器
	// 功能码: 1字节,06 (0x06)
	// address: 2字节,寄存器起始地址,寻址范围[0x0000-0xFFFF]
	// value: 2字节,数据[0x0000-0xFFFF]
	WriteSingleRegister(ctx context.Context, address, value uint16) (results []byte, err error)
	// 读取远程设备中8个内部线圈异常状态(串行线)
	// 功能码: 1字节,07 (0x07)
	ReadExceptionStatus(ctx context.Context) (results []byte, err error)
	// 诊断远程设备(串行线)
	// 功能码: 1字节,08 (0x08)
	// subFunc: 2字节,子功能码[0x0000-0xFFFF]
	// value: N*2字节,数据
	Diagnostics(ctx context.Context, subFunc uint16, value []byte) (results []byte, err error)
	// 读取远程设备的通信事件计数器中的状态字和事件计数(串行线)
	// 功能码: 1字节,11 (0x0B)
	GetCommEventCounter(ctx context.Context) (results []byte, err error)
	// 读取远程设备的状态字、事件计数、消息计数和事件字节的消息字段(串行线)
	// 功能码: 1字节,12 (0x0C)
	GetCommEventLog(ctx context.Context) (results []byte, err error)
	// 在远程设备中强制输出线圈序列中的每个输出线圈的状态(开关量OFF(0)/ON(1))
	// 功能码: 1字节,15 (0x0F)
	// address: 2字节,线圈起始地址,寻址范围[0x0000-0xFFFF]
	// quantity: 2字节,线圈数量[0x0001-0x07B0]
	// value: N*1字节,数据
	WriteMultipleCoils(ctx context.Context, address, quantity uint16, value []byte) (results []byte, err error)
	// 在远程设备中一个连续(1-123)寄存器块写入数据
	// 功能码: 1字节,16 (0x10)
	// address: 2字节,寄存器起始地址,寻址范围[0x0001-0x007B]
	// quantity: 2字节,寄存器数量[0x0001-0x07B0]
	// value: N*2字节,数据
	WriteMultipleregisters(ctx context.Context, address, quantity uint16, value []byte) (results []byte, err error)
}

// ProtocolDataUnit (PDU) is independent of underlying communication layers.
type ProtocolDataUnit struct {
	Code   byte
	Data   []byte
	Refin  bool
	Refout bool
}

// Packager specifies the communication layer.
//
//go:generate mockery -name Packager
type Packager interface {
	Encode(ctx context.Context, pdu *ProtocolDataUnit) (adu []byte, err error)
	Decode(ctx context.Context, adu []byte) (repdu *ProtocolDataUnit, err error)
	Verify(ctx context.Context, adu []byte, readu []byte) (err error)
}

// Transporter specifies the transport layer.
//
//go:generate mockery -name Transporter
type Transporter interface {
	Send(ctx context.Context, adu []byte, waitTimes int, timeout time.Duration) (readu []byte, err error)
}

// 异常状态码常量
const (
	ILLEGAL_FUNCTION                        byte = 1  // 1(0x01) 非法的功能码
	ILLEGAL_DATA_ADDRESS                    byte = 2  // 2(0x02) 非法的数据地址
	ILLEGAL_DATA_VALUE                      byte = 3  // 3(0x03) 非法数据值
	SERVER_DEVICE_FAILURE                   byte = 4  // 4(0x04) 服务器设备故障
	ACKNOWLEDGE                             byte = 5  // 5(0x05) 编程命令相关,服务器已接受请求正在处理中
	SERVER_DEVICE_BUSY                      byte = 6  // 6(0x06) 服务器设备忙
	MEMORY_PARITY_ERROR                     byte = 8  // 8(0x08) 内存奇偶校验错误
	GATEWAY_PATH_UNAVAILABLE                byte = 10 // 10(0x0A) 网关路径不可用
	GATEWAY_TARGET_DEVICE_FAILED_TO_RESPOND byte = 11 // 11(0x0B) 网关目标设备响应失败
)

// 异常处理
type Error struct {
	FunctionCode  byte
	ExceptionCode byte
}

func (e *Error) Error() string {
	var msg string
	switch e.ExceptionCode {
	case ILLEGAL_FUNCTION: // 1(0x01) 非法的功能码
		msg = "illegal function"
	case ILLEGAL_DATA_ADDRESS: // 2(0x02) 非法的数据地址
		msg = "illegal data address"
	case ILLEGAL_DATA_VALUE: // 3(0x03) 非法数据值
		msg = "illegal data value"
	case SERVER_DEVICE_FAILURE: // 4(0x04) 服务器设备故障
		msg = "server device failure"
	case ACKNOWLEDGE: // 5(0x05) 编程命令相关,服务器已接受请求正在处理中
		msg = "acknowledge"
	case SERVER_DEVICE_BUSY: // 6(0x06) 服务器设备忙
		msg = "server device busy"
	case MEMORY_PARITY_ERROR: // 8(0x08) 内存奇偶校验错误
		msg = "memory parity error"
	case GATEWAY_PATH_UNAVAILABLE: // 10(0x0A) 网关路径不可用
		msg = "gateway path unavailable"
	case GATEWAY_TARGET_DEVICE_FAILED_TO_RESPOND: // 11(0x0B) 网关目标设备响应失败
		msg = "gateway target device failed to respond"
	default:
		msg = "unknown"
	}
	return fmt.Sprintf("modbus: exception '%v' (%s), function '%v'", e.ExceptionCode, msg, e.FunctionCode)
}
