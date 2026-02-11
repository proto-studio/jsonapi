package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"proto.zip/studio/jsonapi/pkg/jsonapi"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		http.Error(w, `{"errors":[{"status":"500","title":"Marshal error","detail":""}]}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	_, _ = w.Write(body)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func writeErrors(w http.ResponseWriter, status int, errs []jsonapi.Error) {
	body, err := json.Marshal(jsonapi.ErrorResponse{Errors: errs})
	if err != nil {
		http.Error(w, `{"errors":[{"status":"500","title":"Marshal error","detail":""}]}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	_, _ = w.Write(body)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func selfLink(path string) jsonapi.Links {
	return jsonapi.Links{"self": jsonapi.StringLink(baseURL + path)}
}

func storeToDatum(s StoreRecord, db *DB) jsonapi.Datum[StoreAttributes] {
	d := jsonapi.Datum[StoreAttributes]{
		ID:   s.ID,
		Type: "stores",
		Attributes: StoreAttributes{
			Name:    s.Name,
			Address: s.Address,
		},
		Links: selfLink("/stores/" + s.ID),
		Relationships: map[string]jsonapi.Relationship{
			"pets": {
				Links: jsonapi.Links{
					"self":    jsonapi.StringLink(baseURL + "/stores/" + s.ID + "/relationships/pets"),
					"related": jsonapi.StringLink(baseURL + "/stores/" + s.ID + "/pets"),
				},
				Data: jsonapi.ResourceLinkageCollection(db.petLinkage(s.ID)),
			},
		},
	}
	return d
}

func (db *DB) petLinkage(storeID string) []jsonapi.ResourceIdentifierLinkage {
	pets := db.PetsByStoreID(storeID)
	out := make([]jsonapi.ResourceIdentifierLinkage, len(pets))
	for i, p := range pets {
		out[i] = jsonapi.ResourceIdentifierLinkage{Type: "pets", ID: p.ID}
	}
	return out
}

func petToDatum(p PetRecord, db *DB) jsonapi.Datum[PetAttributes] {
	d := jsonapi.Datum[PetAttributes]{
		ID:   p.ID,
		Type: "pets",
		Attributes: PetAttributes{
			Name:    p.Name,
			Species: p.Species,
		},
		Links: selfLink("/pets/" + p.ID),
	}
	if p.StoreID != "" {
		d.Relationships = map[string]jsonapi.Relationship{
			"store": {
				Links: jsonapi.Links{
					"self":    jsonapi.StringLink(baseURL + "/pets/" + p.ID + "/relationships/store"),
					"related": jsonapi.StringLink(baseURL + "/stores/" + p.StoreID),
				},
				Data: jsonapi.ResourceIdentifierLinkage{Type: "stores", ID: p.StoreID},
			},
		}
	}
	return d
}

func (s *Server) listStores(w http.ResponseWriter, r *http.Request) {
	list := s.db.ListStores()
	data := make([]jsonapi.Datum[StoreAttributes], len(list))
	for i, store := range list {
		data[i] = storeToDatum(store, s.db)
	}
	writeJSON(w, http.StatusOK, jsonapi.DatumCollectionEnvelope[StoreAttributes]{
		Data:  data,
		Links: selfLink("/stores"),
	})
}

func (s *Server) getStore(w http.ResponseWriter, r *http.Request, id string) {
	store, ok := s.db.GetStore(id)
	if !ok {
		writeErrors(w, http.StatusNotFound, []jsonapi.Error{{Status: "404", Title: "Not Found", Detail: "store not found"}})
		return
	}
	writeJSON(w, http.StatusOK, jsonapi.SingleDatumEnvelope[StoreAttributes]{
		Data:  storeToDatum(store, s.db),
		Links: selfLink("/stores/" + id),
	})
}

func (s *Server) createStore(w http.ResponseWriter, r *http.Request) {
	log.Printf("[createStore] entered")
	data, err := readBody(r)
	if err != nil {
		writeErrors(w, http.StatusBadRequest, []jsonapi.Error{{Status: "400", Title: "Bad Request", Detail: err.Error()}})
		return
	}
	ctx := jsonapi.WithMethod(context.Background(), r.Method)
	var env jsonapi.SingleDatumEnvelope[StoreAttributes]
	log.Printf("[createStore] before Apply")
	if errs := StoreRuleSet().Apply(ctx, data, &env); errs != nil {
		log.Printf("[createStore] Apply returned error, converting to JSON:API errors")
		list := jsonapi.ErrorsFromValidationError(errs, jsonapi.SourcePointer)
		log.Printf("[createStore] got %d errors, writing response", len(list))
		writeErrors(w, http.StatusUnprocessableEntity, list)
		log.Printf("[createStore] writeErrors returned")
		return
	}
	log.Printf("[createStore] Apply OK")
	attrs := env.Data.Attributes
	id := nextID(s.db.ListStores(), func(s StoreRecord) string { return s.ID })
	rec := StoreRecord{ID: id, Name: attrs.Name, Address: attrs.Address}
	if err := s.db.CreateStore(rec); err != nil {
		writeErrors(w, http.StatusInternalServerError, []jsonapi.Error{{Status: "500", Title: "Error", Detail: err.Error()}})
		return
	}
	log.Printf("[createStore] created store %s, writing 201", id)
	writeJSON(w, http.StatusCreated, jsonapi.SingleDatumEnvelope[StoreAttributes]{
		Data:  storeToDatum(rec, s.db),
		Links: selfLink("/stores/" + id),
	})
	log.Printf("[createStore] done")
}

func (s *Server) updateStore(w http.ResponseWriter, r *http.Request, id string) {
	data, err := readBody(r)
	if err != nil {
		writeErrors(w, http.StatusBadRequest, []jsonapi.Error{{Status: "400", Title: "Bad Request", Detail: err.Error()}})
		return
	}
	ctx := jsonapi.WithMethod(context.Background(), r.Method)
	ctx = jsonapi.WithId(ctx, id)
	var env jsonapi.SingleDatumEnvelope[StoreAttributes]
	if errs := StoreRuleSet().Apply(ctx, data, &env); errs != nil {
		list := jsonapi.ErrorsFromValidationError(errs, jsonapi.SourcePointer)
		writeErrors(w, http.StatusUnprocessableEntity, list)
		return
	}
	attrs := env.Data.Attributes
	updated, ok := s.db.UpdateStore(id, func(s *StoreRecord) {
		s.Name = attrs.Name
		s.Address = attrs.Address
	})
	if !ok {
		writeErrors(w, http.StatusNotFound, []jsonapi.Error{{Status: "404", Title: "Not Found", Detail: "store not found"}})
		return
	}
	writeJSON(w, http.StatusOK, jsonapi.SingleDatumEnvelope[StoreAttributes]{
		Data:  storeToDatum(updated, s.db),
		Links: selfLink("/stores/" + id),
	})
}

func (s *Server) deleteStore(w http.ResponseWriter, r *http.Request, id string) {
	if !s.db.DeleteStore(id) {
		writeErrors(w, http.StatusNotFound, []jsonapi.Error{{Status: "404", Title: "Not Found", Detail: "store not found"}})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listPets(w http.ResponseWriter, r *http.Request) {
	list := s.db.ListPets()
	data := make([]jsonapi.Datum[PetAttributes], len(list))
	for i, pet := range list {
		data[i] = petToDatum(pet, s.db)
	}
	writeJSON(w, http.StatusOK, jsonapi.DatumCollectionEnvelope[PetAttributes]{
		Data:  data,
		Links: selfLink("/pets"),
	})
}

func (s *Server) getPet(w http.ResponseWriter, r *http.Request, id string) {
	pet, ok := s.db.GetPet(id)
	if !ok {
		writeErrors(w, http.StatusNotFound, []jsonapi.Error{{Status: "404", Title: "Not Found", Detail: "pet not found"}})
		return
	}
	writeJSON(w, http.StatusOK, jsonapi.SingleDatumEnvelope[PetAttributes]{
		Data:  petToDatum(pet, s.db),
		Links: selfLink("/pets/" + id),
	})
}

func (s *Server) createPet(w http.ResponseWriter, r *http.Request) {
	data, err := readBody(r)
	if err != nil {
		writeErrors(w, http.StatusBadRequest, []jsonapi.Error{{Status: "400", Title: "Bad Request", Detail: err.Error()}})
		return
	}
	ctx := jsonapi.WithMethod(context.Background(), r.Method)
	var env jsonapi.SingleDatumEnvelope[PetAttributes]
	if errs := PetRuleSet().Apply(ctx, data, &env); errs != nil {
		list := jsonapi.ErrorsFromValidationError(errs, jsonapi.SourcePointer)
		writeErrors(w, http.StatusUnprocessableEntity, list)
		return
	}
	attrs := env.Data.Attributes
	storeID := ""
	if rel, ok := env.Data.Relationships["store"]; ok {
		if link, ok := rel.Data.(jsonapi.ResourceIdentifierLinkage); ok {
			storeID = link.ID
		}
	}
	id := nextID(s.db.ListPets(), func(p PetRecord) string { return p.ID })
	rec := PetRecord{ID: id, Name: attrs.Name, Species: attrs.Species, StoreID: storeID}
	if err := s.db.CreatePet(rec); err != nil {
		writeErrors(w, http.StatusInternalServerError, []jsonapi.Error{{Status: "500", Title: "Error", Detail: err.Error()}})
		return
	}
	writeJSON(w, http.StatusCreated, jsonapi.SingleDatumEnvelope[PetAttributes]{
		Data:  petToDatum(rec, s.db),
		Links: selfLink("/pets/" + id),
	})
}

func (s *Server) updatePet(w http.ResponseWriter, r *http.Request, id string) {
	data, err := readBody(r)
	if err != nil {
		writeErrors(w, http.StatusBadRequest, []jsonapi.Error{{Status: "400", Title: "Bad Request", Detail: err.Error()}})
		return
	}
	ctx := jsonapi.WithMethod(context.Background(), r.Method)
	ctx = jsonapi.WithId(ctx, id)
	var env jsonapi.SingleDatumEnvelope[PetAttributes]
	if errs := PetRuleSet().Apply(ctx, data, &env); errs != nil {
		list := jsonapi.ErrorsFromValidationError(errs, jsonapi.SourcePointer)
		writeErrors(w, http.StatusUnprocessableEntity, list)
		return
	}
	attrs := env.Data.Attributes
	storeID := ""
	if rel, ok := env.Data.Relationships["store"]; ok {
		if link, ok := rel.Data.(jsonapi.ResourceIdentifierLinkage); ok {
			storeID = link.ID
		}
	}
	updated, ok := s.db.UpdatePet(id, func(p *PetRecord) {
		p.Name = attrs.Name
		p.Species = attrs.Species
		p.StoreID = storeID
	})
	if !ok {
		writeErrors(w, http.StatusNotFound, []jsonapi.Error{{Status: "404", Title: "Not Found", Detail: "pet not found"}})
		return
	}
	writeJSON(w, http.StatusOK, jsonapi.SingleDatumEnvelope[PetAttributes]{
		Data:  petToDatum(updated, s.db),
		Links: selfLink("/pets/" + id),
	})
}

func (s *Server) deletePet(w http.ResponseWriter, r *http.Request, id string) {
	if !s.db.DeletePet(id) {
		writeErrors(w, http.StatusNotFound, []jsonapi.Error{{Status: "404", Title: "Not Found", Detail: "pet not found"}})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func readBody(r *http.Request) (map[string]any, error) {
	var raw map[string]any
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func nextID[T any](list []T, idFn func(T) string) string {
	max := 0
	for _, item := range list {
		id := idFn(item)
		n, _ := strconv.Atoi(id)
		if n > max {
			max = n
		}
	}
	return strconv.Itoa(max + 1)
}
