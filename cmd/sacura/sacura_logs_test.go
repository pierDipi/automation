package main

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParser(t *testing.T) {

	f1, err := NewSingleFileHistoryFinder("testdata/receiver.log")
	if err != nil {
		t.Fatal(err)
	}

	f2, err := NewSingleFileHistoryFinder("testdata/dispatcher.log")
	if err != nil {
		t.Fatal(err)
	}

	parser := SacuraLogParser{
		HistoryFinder: MultiFileHistoryFinder([]HistoryFinder{f1, f2}),
	}

	h, err := parser.Parse("testdata/sacura.log")
	if err != nil {
		t.Fatal(err)
	}

	b, err := json.MarshalIndent(h, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(b))

	expected := History{
		HistoryBySymbol: HistoryBySymbol{
			"+": {
				EventHistory{
					ID: "0002449b-9446-4d47-a346-80d3a00b7b58",
					History: []string{
						`{"l": 1, "id": "0002449b-9446-4d47-a346-80d3a00b7b58"}`,
						`{"l": 3, "id": "0002449b-9446-4d47-a346-80d3a00b7b58"}`,
						`{"l": 7, "id": "0002449b-9446-4d47-a346-80d3a00b7b58"}`,
						`{"l": 9, "id": "0002449b-9446-4d47-a346-80d3a00b7b58"}`,
					},
				},
				EventHistory{ID: "001bb907-bb23-46f8-bafd-25c8cabb5254", History: []string{}},
				{
					ID: "00701efa-b27b-48a2-b448-a91897fdecb6",
					History: []string{
						`{"l": 5, "id": "00701efa-b27b-48a2-b448-a91897fdecb6"}`,
						`{"l": 6, "id": "00701efa-b27b-48a2-b448-a91897fdecb6"}`,
						`{"l": 11, "id": "00701efa-b27b-48a2-b448-a91897fdecb6"}`,
						`{"l": 12, "id": "00701efa-b27b-48a2-b448-a91897fdecb6"}`,
					},
				},
			},
			"-": {
				EventHistory{
					ID: "0010ec9e-bfbc-4643-8647-2fbe49a9a480",
					History: []string{
						`{"l": 2, "id": "0010ec9e-bfbc-4643-8647-2fbe49a9a480"}`,
						`{"l": 4, "id": "0010ec9e-bfbc-4643-8647-2fbe49a9a480"}`,
						`{"l": 8, "id": "0010ec9e-bfbc-4643-8647-2fbe49a9a480"}`,
						`{"l": 10, "id": "0010ec9e-bfbc-4643-8647-2fbe49a9a480"}`,
					},
				},
			},
		},
	}

	if diff := cmp.Diff(expected, h); diff != "" {
		t.Error("(-want, +got) ", diff)
	}
}
