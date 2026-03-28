package service

import (
	"strings"
	"testing"
)

func TestParseCSVBankStatement(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid CSV with English headers",
			input: "date,amount,description,counterparty,reference_num\n" +
				"15.03.2026,1500.50,Оплата бронирования,ООО Рога,12345\n" +
				"16.03.2026,-500.00,Возврат,ИП Иванов,67890\n",
			wantCount: 2,
		},
		{
			name: "valid CSV with Russian headers",
			input: "дата,сумма,описание,контрагент,номер\n" +
				"2026-03-15,2000,Поступление,ООО Тест,\n",
			wantCount: 1,
		},
		{
			name: "amount with comma decimal separator",
			input: "date,amount\n" +
				"15.03.2026,\"1 500,50\"\n",
			wantCount: 1,
		},
		{
			name:    "empty file",
			input:   "date,amount\n",
			wantErr: true,
		},
		{
			name:    "missing required columns",
			input:   "name,description\ntest,test\n",
			wantErr: true,
		},
		{
			name:    "semicolon detected",
			input:   "date;amount;description\n15.03.2026;1500;test\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := ParseCSVBankStatement(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(entries) != tt.wantCount {
				t.Fatalf("expected %d entries, got %d", tt.wantCount, len(entries))
			}
		})
	}
}

func TestParse1CBankStatement(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid 1C format",
			input: "1CClientBankExchange\nВерсияФормата=1.03\n" +
				"СекцияДокумент=Платежное поручение\n" +
				"Номер=123\n" +
				"Дата=15.03.2026\n" +
				"Сумма=1500.50\n" +
				"Плательщик=ООО Рога\n" +
				"НазначениеПлатежа=Оплата бронирования\n" +
				"КонецДокумента\n" +
				"СекцияДокумент=Платежное поручение\n" +
				"Номер=124\n" +
				"Дата=16.03.2026\n" +
				"Сумма=2000\n" +
				"Получатель=ИП Иванов\n" +
				"НазначениеПлатежа=Возврат\n" +
				"КонецДокумента\n",
			wantCount: 2,
		},
		{
			name:    "empty 1C - no sections",
			input:   "1CClientBankExchange\nВерсияФормата=1.03\n",
			wantErr: true,
		},
		{
			name: "1C with missing date",
			input: "СекцияДокумент=Платежное поручение\n" +
				"Сумма=1500\n" +
				"КонецДокумента\n",
			wantErr: true, // will skip unparseable sections, resulting in 0 entries
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := Parse1CBankStatement(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(entries) != tt.wantCount {
				t.Fatalf("expected %d entries, got %d", tt.wantCount, len(entries))
			}
		})
	}
}

func TestParseBankDate(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"15.03.2026", false},
		{"2026-03-15", false},
		{"15/03/2026", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parseBankDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseBankDate(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestParseBankAmount(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"1500.50", 150050, false},
		{"1500,50", 150050, false},
		{"-500.00", -50000, false},
		{"1 500,50", 150050, false},         // thousand separator
		{"1\u00a0500,50", 150050, false},     // non-breaking space
		{"2000", 200000, false},              // no decimals
		{"0", 0, true},                       // zero amount
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseBankAmount(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseBankAmount(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("parseBankAmount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestToBankStatementEntries(t *testing.T) {
	parsed := []ParsedBankEntry{
		{Amount: 150000, Description: "Test 1"},
		{Amount: -50000, Description: "Test 2"},
	}

	entries := ToBankStatementEntries(parsed, [16]byte{1, 2, 3})
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	for _, e := range entries {
		if e.ID == ([16]byte{}) {
			t.Fatal("entry ID should not be zero")
		}
		if e.Status != "pending" {
			t.Fatalf("expected pending status, got %s", e.Status)
		}
	}
}
