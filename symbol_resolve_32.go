//go:build 386
// +build 386

package riptracer

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/ianlancetaylor/demangle"
)

type ELF32_Rela struct {
	R_offset uint32
	R_info   uint32
	R_addend uint32
}

type ELF32_Rel struct {
	R_offset uint32
	R_info   uint32
}

func getType(info uint32) uint8 {
	return uint8(info & 0xFF) // Extracts the upper 8 bits.
}

func getSymb(info uint32) uint32 {
	return info >> 8 // Extracts the sym index
}

func parseELF32RelaEntry(data []byte) (ELF32_Rela, error) {
	var rela ELF32_Rela
	var relaSize = int(unsafe.Sizeof(rela))
	if relaSize > len(data) {
		return rela, nil
	}
	buf := bytes.NewBuffer(data[:relaSize])
	err := binary.Read(buf, binary.LittleEndian, &rela)
	return rela, err
}

func parseELF32RelEntry(data []byte) (ELF32_Rel, error) {
	var rela ELF32_Rel
	var relaSize = int(unsafe.Sizeof(rela))
	if relaSize > len(data) {
		return rela, nil
	}
	buf := bytes.NewBuffer(data[:relaSize])
	err := binary.Read(buf, binary.LittleEndian, &rela)
	return rela, err
}

func parsePlt(f *elf.File, demangleArguments bool) []elf.Symbol {
	addends := true
	plt := make([]elf.Symbol, 0)

	dynSyms, err := f.DynamicSymbols()
	check(err)

	rpSec := f.Section(".rela.plt") //Contain specific addends
	if nil == rpSec {
		addends = false
		rpSec = f.Section(".rel.plt") //Do not contain addends...
	}

	cnt := 0
	data, err := rpSec.Data()
	check(err)

	for cnt = 0; cnt < int(rpSec.Size); cnt += int(rpSec.Entsize) {
		var idx uint32
		var info uint32

		if addends {
			rela, err := parseELF32RelaEntry(data[cnt:])
			if err != nil {
				break
			}
			info = rela.R_info
		} else {
			rel, err := parseELF32RelEntry(data[cnt:])
			if err != nil {
				break
			}
			info = rel.R_info
		}
		idx = getSymb(info) - 1
		sym := dynSyms[idx]
		var demangledName string
		if demangleArguments {
			demangledName, err = demangle.ToString(sym.Name, demangle.Option(demangle.NoTemplateParams), demangle.Option(demangle.LLVMStyle))
		} else {
			demangledName, err = demangle.ToString(sym.Name, demangle.Option(demangle.NoParams), demangle.Option(demangle.NoTemplateParams), demangle.Option(demangle.LLVMStyle))
		}

		if err != nil {
			demangledName = sym.Name
		}
		sym.Name = demangledName
		//fmt.Printf("name: %v\n", sym.Name)
		plt = append(plt, sym)
	}
	return plt
}

type SymbolResolver struct {
	PLT               []elf.Symbol
	pltSection        *elf.Section
	demangleArguments bool
}

func NewSymbolResolver(filepath string, demangleArguments bool) (*SymbolResolver, error) {
	f, err := elf.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pltSect := f.Section(".plt")
	if pltSect == nil {
		return nil, fmt.Errorf("couldn't find dynstr")
	}

	s := SymbolResolver{pltSection: pltSect, demangleArguments: demangleArguments}
	s.PLT = parsePlt(f, demangleArguments)

	return &s, nil
}

func (s *SymbolResolver) GetPLTOffsetBySymName(symName string) (uintptr, error) {
	for i := range s.PLT {
		sym := s.PLT[i]
		if sym.Name == symName {
			//fmt.Printf("Found name: %v, addr: 0x%x, entsize: %d, i: %d\n", sym.Name, s.pltSection.Addr, s.pltSection.Entsize, i)
			//addrOffset := s.pltSection.Addr + s.pltSection.Entsize + (uint64(i) * s.pltSection.Entsize)
			addrOffset := s.pltSection.Addr + uint64(0x10) + (uint64(i) * 0x10)
			return uintptr(addrOffset), nil
		}
	}

	return 0, fmt.Errorf("Couldn't find symName in file")
}

func (s *SymbolResolver) GetPLTSymNameByOffset(offset uint64) (string, error) {
	for i := range s.PLT {
		//addrOffset := s.pltSection.Addr + s.pltSection.Entsize + (uint64(i) * s.pltSection.Entsize)
		addrOffset := s.pltSection.Addr + uint64(0x10) + (uint64(i) * 0x10)
		if offset == addrOffset {
			sym := s.PLT[i]
			return sym.Name, nil
		}
	}

	return "", fmt.Errorf("Couldn't find symbol at offset 0x%8.8x", offset)
}

func (s *SymbolResolver) PrintSymNames() {
	for i := range s.PLT {
		sym := s.PLT[i]
		addrOffset := s.pltSection.Addr + uint64(0x10) + (uint64(i) * 0x10)
		fmt.Printf("Found name: %v, addr: 0x%x\n", sym.Name, addrOffset)
		//addrOffset := s.pltSection.Addr + s.pltSection.Entsize + (uint64(i) * s.pltSection.Entsize)
	}
	return
}
