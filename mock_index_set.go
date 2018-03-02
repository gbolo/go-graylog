package graylog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// AddIndexSet adds a index set to the Mock Server.
func (ms *MockServer) AddIndexSet(indexSet *IndexSet) {
	if indexSet.Id == "" {
		indexSet.Id = randStringBytesMaskImprSrc(24)
	}
	ms.IndexSets[indexSet.Id] = *indexSet
	ms.safeSave()
}

// DeleteIndexSet removes a index set from the Mock Server.
func (ms *MockServer) DeleteIndexSet(id string) {
	delete(ms.IndexSets, id)
	// delete(ms.IndexSetStats, id)
	ms.safeSave()
}

// IndexSetList returns a list of all index sets.
func (ms *MockServer) IndexSetList() []IndexSet {
	if ms.IndexSets == nil {
		return []IndexSet{}
	}
	arr := make([]IndexSet, len(ms.IndexSets))
	i := 0
	for _, index := range ms.IndexSets {
		arr[i] = index
		i++
	}
	return arr
}

func validateIndexSet(indexSet *IndexSet) (int, []byte) {
	if indexSet.Title == "" {
		return 400, []byte(`{"type": "ApiError", "message": "Can not construct instance of org.graylog2.rest.resources.system.indexer.responses.IndexSetSummary, problem: Null title\n at [Source: org.glassfish.jersey.message.internal.ReaderInterceptorExecutor$UnCloseableInputStream@43956bc7; line: 1, column: 2]"}`)
	}
	if indexSet.IndexPrefix == "" {
		return 400, []byte(`{"type": "ApiError", "message": "Can not construct instance of org.graylog2.rest.resources.system.indexer.responses.IndexSetSummary, problem: Null indexPrefix\n at [Source: org.glassfish.jersey.message.internal.ReaderInterceptorExecutor$UnCloseableInputStream@637e3792; line: 1, column: 17]"}`)
	}
	if indexSet.RotationStrategyClass == "" {
		return 400, []byte(`{"type": "ApiError", "message": "Can not construct instance of org.graylog2.rest.resources.system.indexer.responses.IndexSetSummary, problem: Null rotationStrategyClass\n at [Source: org.glassfish.jersey.message.internal.ReaderInterceptorExecutor$UnCloseableInputStream@5e366094; line: 1, column: 41]"}`)
	}
	if indexSet.RotationStrategy == nil {
		return 400, []byte(`{"type": "ApiError", "message": "Can not construct instance of org.graylog2.rest.resources.system.indexer.responses.IndexSetSummary, problem: Null rotationStrategy\n at [Source: org.glassfish.jersey.message.internal.ReaderInterceptorExecutor$UnCloseableInputStream@12f1391d; line: 1, column: 141]"}`)
	}
	return 200, []byte("")
}

// GET /system/indices/index_sets Get a list of all index sets
func (ms *MockServer) handleGetIndexSets(
	w http.ResponseWriter, r *http.Request, _ httprouter.Params,
) {
	ms.Logger.WithFields(log.Fields{
		"path": r.URL.Path, "method": r.Method,
	}).Info("request start")
	w.Header().Set("Content-Type", "application/json")
	arr := ms.IndexSetList()
	indexSets := indexSetsBody{
		IndexSets: arr, Total: len(arr), Stats: &IndexSetStats{}}
	b, err := json.Marshal(&indexSets)
	if err != nil {
		write500Error(w)
	} else {
		w.Write(b)
	}
}

// GET /system/indices/index_sets/{id} Get index set
func (ms *MockServer) handleGetIndexSet(
	w http.ResponseWriter, r *http.Request, ps httprouter.Params,
) {
	id := ps.ByName("indexSetId")
	if id == "stats" {
		ms.handleGetAllIndexSetsStats(w, r, ps)
		return
	}
	ms.Logger.WithFields(log.Fields{
		"path": r.URL.Path, "method": r.Method,
	}).Info("request start")
	w.Header().Set("Content-Type", "application/json")
	indexSet, ok := ms.IndexSets[id]
	if !ok {
		w.WriteHeader(404)
		w.Write([]byte(fmt.Sprintf(
			`{"type": "ApiError", "message": "No indexSet found with id %s"}`, id)))
		return
	}
	b, err := json.Marshal(&indexSet)
	if err != nil {
		write500Error(w)
	} else {
		w.Write(b)
	}
}

// POST /system/indices/index_sets Create index set
func (ms *MockServer) handleCreateIndexSet(
	w http.ResponseWriter, r *http.Request, _ httprouter.Params,
) {
	ms.Logger.WithFields(log.Fields{
		"path": r.URL.Path, "method": r.Method,
	}).Info("request start")
	w.Header().Set("Content-Type", "application/json")
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		write500Error(w)
		return
	}
	indexSet := &IndexSet{}
	err = json.Unmarshal(b, indexSet)
	if err != nil {
		ms.Logger.WithFields(log.Fields{
			"body": string(b), "error": err,
		}).Info("Failed to parse request body as IndexSet")
		w.WriteHeader(400)
		w.Write([]byte(`{"message":"400 Bad Request"}`))
		return
	}
	ms.Logger.WithFields(log.Fields{
		"body": string(b), "index_set": indexSet,
	}).Debug("request body")
	sc, msg := validateIndexSet(indexSet)
	if sc != 200 {
		w.WriteHeader(sc)
		w.Write(msg)
		return
	}
	ms.AddIndexSet(indexSet)
	b, err = json.Marshal(indexSet)
	if err != nil {
		write500Error(w)
	} else {
		w.Write(b)
	}
}

// PUT /system/indices/index_sets/{id} Update index set
func (ms *MockServer) handleUpdateIndexSet(
	w http.ResponseWriter, r *http.Request, ps httprouter.Params,
) {
	ms.Logger.WithFields(log.Fields{
		"path": r.URL.Path, "method": r.Method,
	}).Info("request start")
	w.Header().Set("Content-Type", "application/json")
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		write500Error(w)
		return
	}
	id := ps.ByName("indexSetId")
	if _, ok := ms.IndexSets[id]; !ok {
		w.WriteHeader(404)
		w.Write([]byte(fmt.Sprintf(
			`{"type": "ApiError", "message": "No indexSet found with id %s"}`, id)))
		return
	}
	indexSet := &IndexSet{}
	err = json.Unmarshal(b, indexSet)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(`{"message":"400 Bad Request"}`))
		return
	}
	indexSet.Id = id
	sc, msg := validateIndexSet(indexSet)
	if sc != 200 {
		w.WriteHeader(sc)
		w.Write(msg)
		return
	}
	ms.AddIndexSet(indexSet)
	b, err = json.Marshal(indexSet)
	if err != nil {
		write500Error(w)
	} else {
		w.Write(b)
	}
}

// DELETE /system/indices/index_sets/{id} Delete index set
func (ms *MockServer) handleDeleteIndexSet(
	w http.ResponseWriter, r *http.Request, ps httprouter.Params,
) {
	ms.Logger.WithFields(log.Fields{
		"path": r.URL.Path, "method": r.Method,
	}).Info("request start")
	w.Header().Set("Content-Type", "application/json")
	id := ps.ByName("indexSetId")
	_, ok := ms.IndexSets[id]
	if !ok {
		w.WriteHeader(404)
		w.Write([]byte(fmt.Sprintf(
			`{"type": "ApiError", "message": "No indexSet found with id %s"}`, id)))
		return
	}
	ms.DeleteIndexSet(id)
}

// PUT /system/indices/index_sets/{id}/default Set default index set
func (ms *MockServer) handleSetDefaultIndexSet(
	w http.ResponseWriter, r *http.Request, ps httprouter.Params,
) {
	ms.Logger.WithFields(log.Fields{
		"path": r.URL.Path, "method": r.Method,
	}).Info("request start")
	w.Header().Set("Content-Type", "application/json")
	id := ps.ByName("indexSetId")
	indexSet, ok := ms.IndexSets[id]
	if !ok {
		w.WriteHeader(404)
		w.Write([]byte(fmt.Sprintf(
			`{"type": "ApiError", "message": "No indexSet found with id %s"}`, id)))
		return
	}
	if !indexSet.Writable {
		w.WriteHeader(409)
		w.Write([]byte(
			`{"type": "ApiError", "message": "Default index set must be writable."}`))
		return
	}
	for k, v := range ms.IndexSets {
		if v.Default {
			v.Default = false
			ms.IndexSets[k] = v
			break
		}
	}
	indexSet.Default = true
	ms.AddIndexSet(&indexSet)
	b, err := json.Marshal(&indexSet)
	if err != nil {
		write500Error(w)
	} else {
		w.Write(b)
	}
}

// GET /system/indices/index_sets/{id}/stats Get index set statistics
func (ms *MockServer) handleGetIndexSetStats(
	w http.ResponseWriter, r *http.Request, ps httprouter.Params,
) {
	ms.Logger.WithFields(log.Fields{
		"path": r.URL.Path, "method": r.Method,
	}).Info("request start")
	w.Header().Set("Content-Type", "application/json")
	id := ps.ByName("indexSetId")
	indexSetStats, ok := ms.IndexSetStats[id]
	if !ok {
		w.WriteHeader(404)
		w.Write([]byte(fmt.Sprintf(
			`{"type": "ApiError", "message": "No indexSet found with id %s"}`, id)))
		return
	}
	b, err := json.Marshal(&indexSetStats)
	if err != nil {
		write500Error(w)
	} else {
		w.Write(b)
	}
}

// GET /system/indices/index_sets/stats Get stats of all index sets
func (ms *MockServer) handleGetAllIndexSetsStats(
	w http.ResponseWriter, r *http.Request, ps httprouter.Params,
) {
	ms.Logger.WithFields(log.Fields{
		"path": r.URL.Path, "method": r.Method,
	}).Info("request start")
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(ms.AllIndexSetsStats())
	if err != nil {
		write500Error(w)
	} else {
		w.Write(b)
	}
}
