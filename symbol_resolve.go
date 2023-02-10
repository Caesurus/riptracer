package riptracer

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/ianlancetaylor/demangle"
)

type ELF64_Rela_Info struct {
	Type uint32
	Sym  uint32
}

type ELF64_Rela struct {
	R_offset uint64
	R_info   ELF64_Rela_Info
	R_addend int64
}

type ELF32_Rela_Info struct {
	Type uint32
	Sym  uint32
}

type ELF32_Rela struct {
	R_offset uint32
	R_info   ELF32_Rela_Info
	R_addend int32
}

func parseELF64RelaEntry(data []byte) (ELF64_Rela, error) {
	var rela ELF64_Rela
	var relaSize = int(unsafe.Sizeof(rela))
	buf := bytes.NewBuffer(data[:relaSize])
	err := binary.Read(buf, binary.LittleEndian, &rela)
	return rela, err
}
func parsePlt(f *elf.File) []elf.Symbol {
	plt := make([]elf.Symbol, 0)

	dynSyms, err := f.DynamicSymbols()
	check(err)

	rpSec := f.Section(".rela.plt")
	cnt := 0
	data, err := rpSec.Data()
	check(err)

	for cnt = 0; cnt < int(rpSec.Size); cnt += int(rpSec.Entsize) {
		rela, err := parseELF64RelaEntry(data[cnt:])
		if err != nil {
			break
		}

		idx := rela.R_info.Sym - 1
		sym := dynSyms[idx]
		demangledName, err := demangle.ToString(sym.Name, demangle.Option(demangle.NoParams), demangle.Option(demangle.NoTemplateParams), demangle.Option(demangle.LLVMStyle))
		if err != nil {
			demangledName = sym.Name
		}
		sym.Name = demangledName
		plt = append(plt, sym)
	}
	return plt
}

type SymbolResolver struct {
	elfFile    *elf.File
	PLT        []elf.Symbol
	pltSection *elf.Section
}

func NewSymbolResolver(filepath string) (*SymbolResolver, error) {
	f, err := elf.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pltSect := f.Section(".plt")
	if pltSect == nil {
		return nil, fmt.Errorf("Couldn't find dynstr")
	}

	s := SymbolResolver{elfFile: f, pltSection: pltSect}
	s.PLT = parsePlt(f)
	return &s, nil
}

func (s *SymbolResolver) GetPLTOffsetBySymName(symName string) (uintptr, error) {
	for i := range s.PLT {
		sym := s.PLT[i]
		if sym.Name == symName {
			addrOffset := s.pltSection.Addr + s.pltSection.Entsize + (uint64(i) * s.pltSection.Entsize)
			return uintptr(addrOffset), nil
		}
	}

	return 0, fmt.Errorf("Couldn't find symName in file")
}

func (s *SymbolResolver) GetPLTSymNameByOffset(offset uint64) (string, error) {

	/*
		idx := (offset - (s.pltSection.Addr + s.pltSection.Entsize)) / s.pltSection.Entsize

		if idx <= uint64(len(s.PLT)) {
			sym := s.PLT[idx]
			return sym.Name, nil
		}
	*/
	for i := range s.PLT {
		addrOffset := s.pltSection.Addr + s.pltSection.Entsize + (uint64(i) * s.pltSection.Entsize)
		if offset == addrOffset {
			sym := s.PLT[i]
			return sym.Name, nil
		}
	}

	return "", fmt.Errorf("Couldn't find symbol at offset 0x%8.8x", offset)
}
