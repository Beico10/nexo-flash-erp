package dispatch

import (
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ParseXMLBatch processa arquivos XML de NF-e.
func ParseXMLBatch(data []byte) *ParseResult {
	result := &ParseResult{Items: []*DeliveryItem{}}

	// Estrutura simplificada para NF-e
	type NFe struct {
		InfNFe struct {
			ID  string `xml:"Id,attr"`
			Ide struct {
				NNF string `xml:"nNF"`
			} `xml:"ide"`
			Dest struct {
				CNPJCPF string `xml:"CNPJ"`
				CPF     string `xml:"CPF"`
				XNome   string `xml:"xNome"`
				EnderDest struct {
					XLgr    string `xml:"xLgr"`
					Nro     string `xml:"nro"`
					XBairro string `xml:"xBairro"`
					XMun    string `xml:"xMun"`
					UF      string `xml:"UF"`
					CEP     string `xml:"CEP"`
				} `xml:"enderDest"`
			} `xml:"dest"`
			Transp struct {
				Vol []struct {
					QVol  string `xml:"qVol"`
					PesoL string `xml:"pesoL"`
				} `xml:"vol"`
			} `xml:"transp"`
			Total struct {
				ICMSTot struct {
					VNF string `xml:"vNF"`
				} `xml:"ICMSTot"`
			} `xml:"total"`
		} `xml:"infNFe"`
	}

	type NFeProc struct {
		NFe NFe `xml:"NFe"`
	}

	// Tentar parsear como nfeProc ou NFe direta
	var proc NFeProc
	if err := xml.Unmarshal(data, &proc); err != nil {
		// Tentar como NFe direta
		var nfe NFe
		if err2 := xml.Unmarshal(data, &nfe); err2 != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("XML inválido: %v", err))
			result.Failed++
			return result
		}
		proc.NFe = nfe
	}

	nfe := proc.NFe.InfNFe
	if nfe.ID == "" && nfe.Ide.NNF == "" {
		result.Errors = append(result.Errors, "NF-e sem identificação")
		result.Failed++
		return result
	}

	// Montar item
	doc := nfe.Dest.CNPJCPF
	if doc == "" {
		doc = nfe.Dest.CPF
	}

	addr := fmt.Sprintf("%s, %s - %s, %s/%s",
		nfe.Dest.EnderDest.XLgr,
		nfe.Dest.EnderDest.Nro,
		nfe.Dest.EnderDest.XBairro,
		nfe.Dest.EnderDest.XMun,
		nfe.Dest.EnderDest.UF,
	)

	var totalWeight float64
	var totalVolumes int
	for _, vol := range nfe.Transp.Vol {
		if w, err := strconv.ParseFloat(vol.PesoL, 64); err == nil {
			totalWeight += w
		}
		if q, err := strconv.Atoi(vol.QVol); err == nil {
			totalVolumes += q
		}
	}
	if totalVolumes == 0 {
		totalVolumes = 1
	}

	totalValue, _ := strconv.ParseFloat(nfe.Total.ICMSTot.VNF, 64)

	item := &DeliveryItem{
		ID:            uuid.New().String(),
		NFeKey:        strings.TrimPrefix(nfe.ID, "NFe"),
		NFeNumber:     nfe.Ide.NNF,
		DocumentType:  "nfe",
		RecipientName: nfe.Dest.XNome,
		RecipientDoc:  doc,
		FullAddress:   addr,
		City:          nfe.Dest.EnderDest.XMun,
		State:         nfe.Dest.EnderDest.UF,
		ZipCode:       nfe.Dest.EnderDest.CEP,
		GeoStatus:     "pending",
		WeightKg:      totalWeight,
		Volumes:       totalVolumes,
		TotalValue:    totalValue,
		Status:        "pending",
		ImportSource:  "xml",
		ImportedAt:    time.Now(),
		Priority:      2,
	}

	result.Items = append(result.Items, item)
	result.Processed++
	return result
}

// ParseCSVBatch processa arquivos CSV.
func ParseCSVBatch(data []byte) *ParseResult {
	result := &ParseResult{Items: []*DeliveryItem{}}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = ';'
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		// Tentar com vírgula
		reader = csv.NewReader(bytes.NewReader(data))
		reader.Comma = ','
		reader.LazyQuotes = true
		records, err = reader.ReadAll()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("CSV inválido: %v", err))
			return result
		}
	}

	if len(records) < 2 {
		result.Errors = append(result.Errors, "CSV vazio ou sem dados")
		return result
	}

	// Mapear colunas
	header := records[0]
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	getCol := func(row []string, names ...string) string {
		for _, name := range names {
			if idx, ok := colMap[name]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
		}
		return ""
	}

	for i, row := range records[1:] {
		if len(row) < 3 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Linha %d ignorada: poucos campos", i+2))
			continue
		}

		name := getCol(row, "destinatario", "nome", "cliente", "recipient", "name")
		if name == "" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Linha %d: sem destinatário", i+2))
			result.Failed++
			continue
		}

		weight, _ := strconv.ParseFloat(strings.Replace(getCol(row, "peso", "weight", "peso_kg"), ",", ".", 1), 64)
		value, _ := strconv.ParseFloat(strings.Replace(getCol(row, "valor", "value", "total"), ",", ".", 1), 64)
		volumes, _ := strconv.Atoi(getCol(row, "volumes", "qtd", "quantidade"))
		if volumes == 0 {
			volumes = 1
		}

		item := &DeliveryItem{
			ID:            uuid.New().String(),
			NFeKey:        getCol(row, "chave", "nfe_key", "chave_nfe"),
			NFeNumber:     getCol(row, "numero", "nfe", "nota"),
			DocumentType:  "nfe",
			RecipientName: name,
			RecipientDoc:  getCol(row, "cpf", "cnpj", "documento", "doc"),
			FullAddress:   getCol(row, "endereco", "address", "logradouro"),
			City:          getCol(row, "cidade", "city", "municipio"),
			State:         getCol(row, "uf", "estado", "state"),
			ZipCode:       getCol(row, "cep", "zipcode"),
			GeoStatus:     "pending",
			WeightKg:      weight,
			Volumes:       volumes,
			TotalValue:    value,
			Status:        "pending",
			ImportSource:  "csv",
			ImportedAt:    time.Now(),
			Priority:      2,
		}

		result.Items = append(result.Items, item)
		result.Processed++
	}

	return result
}

// ParseEDIBatch processa arquivos EDI/X12.
func ParseEDIBatch(data []byte) *ParseResult {
	result := &ParseResult{Items: []*DeliveryItem{}}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Formato simples: CHAVE;DEST;CIDADE;UF;PESO;VOLUMES;VALOR
		parts := strings.Split(line, ";")
		if len(parts) < 4 {
			continue
		}

		weight, _ := strconv.ParseFloat(parts[4], 64)
		volumes := 1
		if len(parts) > 5 {
			volumes, _ = strconv.Atoi(parts[5])
		}
		value := 0.0
		if len(parts) > 6 {
			value, _ = strconv.ParseFloat(parts[6], 64)
		}

		item := &DeliveryItem{
			ID:            uuid.New().String(),
			NFeKey:        parts[0],
			DocumentType:  "edi",
			RecipientName: parts[1],
			City:          parts[2],
			State:         parts[3],
			GeoStatus:     "pending",
			WeightKg:      weight,
			Volumes:       volumes,
			TotalValue:    value,
			Status:        "pending",
			ImportSource:  "edi",
			ImportedAt:    time.Now(),
			Priority:      2,
		}

		result.Items = append(result.Items, item)
		result.Processed++
	}

	if result.Processed == 0 {
		result.Errors = append(result.Errors, "Nenhum registro EDI válido encontrado")
	}

	return result
}

// DistributeByVehicle distribui entregas entre veículos.
func DistributeByVehicle(items []*DeliveryItem, vehicles []*Vehicle) []*VehicleAssignment {
	assignments := make([]*VehicleAssignment, len(vehicles))
	for i, v := range vehicles {
		assignments[i] = &VehicleAssignment{
			Vehicle: v,
			Items:   []*DeliveryItem{},
		}
	}

	// Distribuição round-robin simples considerando capacidade
	vIdx := 0
	for _, item := range items {
		// Encontrar veículo com capacidade
		tried := 0
		for tried < len(vehicles) {
			a := assignments[vIdx]
			v := a.Vehicle

			canFitWeight := v.MaxWeightKg == 0 || a.TotalWeight+item.WeightKg <= v.MaxWeightKg
			canFitStops := v.MaxStops == 0 || a.TotalStops < v.MaxStops

			if canFitWeight && canFitStops {
				a.Items = append(a.Items, item)
				a.TotalStops++
				a.TotalWeight += item.WeightKg
				a.TotalCubage += item.CubageM3
				item.VehicleID = v.ID
				break
			}

			vIdx = (vIdx + 1) % len(vehicles)
			tried++
		}

		// Se nenhum coube, força no atual
		if tried >= len(vehicles) {
			a := assignments[vIdx]
			a.Items = append(a.Items, item)
			a.TotalStops++
			a.TotalWeight += item.WeightKg
			a.TotalCubage += item.CubageM3
			item.VehicleID = a.Vehicle.ID
		}

		vIdx = (vIdx + 1) % len(vehicles)
	}

	// Calcular percentuais
	for _, a := range assignments {
		if a.Vehicle.MaxWeightKg > 0 {
			a.WeightUsed = (a.TotalWeight / a.Vehicle.MaxWeightKg) * 100
		}
		if a.Vehicle.MaxCubageM3 > 0 {
			a.CubageUsed = (a.TotalCubage / a.Vehicle.MaxCubageM3) * 100
		}
	}

	return assignments
}
