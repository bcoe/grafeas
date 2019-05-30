// Copyright 2017 The Grafeas Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// package v1alpha1 is an implementation of the v1alpha1 version of Grafeas.
package v1alpha1

import (
	"fmt"
	"log"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	pb "github.com/grafeas/grafeas/proto/v1/grafeas_go_proto"
	prpb "github.com/grafeas/grafeas/proto/v1/project_go_proto"
	"github.com/grafeas/grafeas/samples/server/go-server/api/server/name"
	"github.com/grafeas/grafeas/server-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Grafeas is an implementation of the Grafeas API, which should be called by handler methods for verification of logic
// and storage.
type Grafeas struct {
	S server.Storager
}

// maxBatch is the maximum data size in the batch API according to the protocol specification
const maxBatch = 1000

func (g *Grafeas) GetVulnerabilityOccurrencesSummary(ctx context.Context, req *pb.GetVulnerabilityOccurrencesSummaryRequest) (*pb.VulnerabilityOccurrencesSummary, error) {
	return nil, status.Error(codes.Unimplemented, "GetVulnerabilityOccurrencesSummary is not yet implemented")
}

// BatchCreateNotes validates that all notes are valid and then add them to the backing datastore.
func (g *Grafeas) BatchCreateNotes(ctx context.Context, in *pb.BatchCreateNotesRequest) (*pb.BatchCreateNotesResponse, error) {
	if len(in.Notes) > maxBatch {
		log.Printf("Too many notes in batch %d", len(in.Notes))
		return nil, status.Error(codes.InvalidArgument, "Too many notes")
	}
	var resp pb.BatchCreateNotesResponse
	for _, n := range in.Notes {
		note, err := g.createNote(ctx, n)
		if err != nil {
			return nil, err
		}
		resp.Notes = append(resp.Notes, note)
	}
	return &resp, nil
}

// BatchCreateOccurrences validates that all notes are valid and then creates an associated occurrence per note in the backing datastore.
func (g *Grafeas) BatchCreateOccurrences(ctx context.Context, req *pb.BatchCreateOccurrencesRequest) (*pb.BatchCreateOccurrencesResponse, error) {
	if len(req.Occurrences) > maxBatch {
		log.Printf("Too many occurences in batch %d", len(req.Occurrences))
		return nil, status.Error(codes.InvalidArgument, "Too many occurrences")
	}
	var resp pb.BatchCreateOccurrencesResponse
	for _, o := range req.Occurrences {
		occ, err := g.createOccurrence(ctx, o, req.Parent)
		if err != nil {
			return nil, err
		}
		resp.Occurrences = append(resp.Occurrences, occ)
	}
	return &resp, nil
}

// CreateProject validates that a project is valid and then creates a project in the backing datastore.
func (g *Grafeas) CreateProject(ctx context.Context, req *prpb.CreateProjectRequest) (*prpb.Project, error) {
	p := req.Project
	if req == nil {
		log.Print("Project must not be empty.")
		return nil, status.Error(codes.InvalidArgument, "Project must not be empty")
	}
	if p.Name == "" {
		log.Printf("Project name must not be empty: %v", p.Name)
		return nil, status.Error(codes.InvalidArgument, "Project name must not be empty")
	}
	pID, err := name.ParseProject(p.Name)
	if err != nil {
		log.Printf("Invalid project name: %v", p.Name)
		return nil, status.Error(codes.InvalidArgument, "Invalid project name")
	}
	return p, g.S.CreateProject(pID)
}

// CreateNote validates that a note is valid and then creates a note in the backing datastore.
func (g *Grafeas) CreateNote(ctx context.Context, req *pb.CreateNoteRequest) (*pb.Note, error) {
	return g.createNote(ctx, req.Note)
}

// CreateOccurrence validates that a note is valid and then creates an occurrence in the backing datastore.
func (g *Grafeas) CreateOccurrence(ctx context.Context, req *pb.CreateOccurrenceRequest) (*pb.Occurrence, error) {
	return g.createOccurrence(ctx, req.Occurrence, req.Parent)
}

// DeleteProject deletes a project from the datastore.
func (g *Grafeas) DeleteProject(ctx context.Context, req *prpb.DeleteProjectRequest) (*empty.Empty, error) {
	pID, err := name.ParseProject(req.Name)
	if err != nil {
		log.Printf("Error parsing project name: %v", req.Name)
		return nil, status.Error(codes.InvalidArgument, "Invalid Project name")
	}
	return &empty.Empty{}, g.S.DeleteProject(pID)
}

// DeleteOccurrence deletes an occurrence from the datastore.
func (g *Grafeas) DeleteOccurrence(ctx context.Context, req *pb.DeleteOccurrenceRequest) (*empty.Empty, error) {
	pID, oID, err := name.ParseOccurrence(req.Name)
	if err != nil {
		log.Printf("Error parsing name: %v", req.Name)
		return nil, status.Error(codes.InvalidArgument, "Invalid occurrence name")
	}
	return &empty.Empty{}, g.S.DeleteOccurrence(pID, oID)
}

// DeleteNote deletes a note from the datastore.
func (g *Grafeas) DeleteNote(ctx context.Context, req *pb.DeleteNoteRequest) (*empty.Empty, error) {
	pID, nID, err := name.ParseNote(req.Name)
	if err != nil {
		log.Printf("Error parsing name: %v", req.Name)
		return nil, status.Error(codes.InvalidArgument, "Invalid note name")
	}
	// TODO: Check for occurrences tied to this note, and return an error if there are any before deletion.
	return &empty.Empty{}, g.S.DeleteNote(pID, nID)
}

// GetProject gets a project from the datastore.
func (g *Grafeas) GetProject(ctx context.Context, req *prpb.GetProjectRequest) (*prpb.Project, error) {
	pID, err := name.ParseProject(req.Name)
	if err != nil {
		log.Printf("Error parsing project name: %v", req.Name)
		return nil, status.Error(codes.InvalidArgument, "Invalid Project name")
	}
	return g.S.GetProject(pID)
}

// GetNote gets a note from the datastore.
func (g *Grafeas) GetNote(ctx context.Context, req *pb.GetNoteRequest) (*pb.Note, error) {
	pID, nID, err := name.ParseNote(req.Name)
	if err != nil {
		log.Printf("Error parsing name: %v", req.Name)
		return nil, status.Error(codes.InvalidArgument, "Invalid Note name")
	}
	return g.S.GetNote(pID, nID)
}

// GetOccurrence gets a occurrence from the datastore.
func (g *Grafeas) GetOccurrence(ctx context.Context, req *pb.GetOccurrenceRequest) (*pb.Occurrence, error) {
	pID, oID, err := name.ParseOccurrence(req.Name)
	if err != nil {
		log.Printf("Could note parse name %v", req.Name)
		return nil, status.Error(codes.InvalidArgument, "Could note parse name")
	}
	return g.S.GetOccurrence(pID, oID)
}

// GetOccurrenceNote gets a the note for the provided occurrence from the datastore.
func (g *Grafeas) GetOccurrenceNote(ctx context.Context, req *pb.GetOccurrenceNoteRequest) (*pb.Note, error) {
	pID, oID, err := name.ParseOccurrence(req.Name)
	if err != nil {
		log.Printf("Error parsing name: %v", req.Name)
		return nil, status.Error(codes.InvalidArgument, "Invalid occurrence name")
	}
	o, gErr := g.S.GetOccurrence(pID, oID)
	if gErr != nil {
		return nil, gErr
	}
	npID, nID, err := name.ParseNote(o.NoteName)
	if err != nil {
		log.Printf("Invalid note name: %v", o.Name)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid note name: %v", o.NoteName))
	}
	return g.S.GetNote(npID, nID)
}

func (g *Grafeas) UpdateNote(ctx context.Context, req *pb.UpdateNoteRequest) (*pb.Note, error) {
	pID, nID, err := name.ParseNote(req.Name)
	if err != nil {
		log.Printf("Error parsing name: %v", req.Name)
		return nil, status.Error(codes.InvalidArgument, "Invalid Note name")
	}
	// get existing note
	existing, gErr := g.S.GetNote(pID, nID)
	if gErr != nil {
		return nil, err
	}
	// verify that name didnt change
	if req.Note.Name != existing.Name {
		log.Printf("Cannot change note name: %v", req.Note.Name)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Cannot change note name: %v", req.Note.Name))
	}

	// update note
	if gErr = g.S.UpdateNote(pID, nID, req.Note); err != nil {
		log.Printf("Cannot update note : %v", gErr)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Cannot change note name: %v", req.Note.Name))
	}
	return g.S.GetNote(pID, nID)
}

func (g *Grafeas) UpdateOccurrence(ctx context.Context, req *pb.UpdateOccurrenceRequest) (*pb.Occurrence, error) {
	pID, oID, err := name.ParseOccurrence(req.Name)
	if err != nil {
		log.Printf("Error parsing name: %v", req.Name)
		return nil, status.Error(codes.InvalidArgument, "Invalid occurrence name")
	}
	// get existing Occurrence
	existing, gErr := g.S.GetOccurrence(pID, oID)
	if gErr != nil {
		return nil, gErr
	}

	// verify that name did not change
	if req.Name != existing.Name {
		log.Printf("Cannot change occurrence name: %v", req.Occurrence.Name)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Cannot change occurrence name: %v", req.Occurrence.Name))
	}
	// verify that if note name changed, it still exists
	if req.Occurrence.NoteName != existing.NoteName {
		npID, nID, err := name.ParseNote(req.Occurrence.NoteName)
		if err != nil {
			return nil, err
		}
		if newN, err := g.S.GetNote(npID, nID); newN == nil || err != nil {
			return nil, err
		}
	}

	// update Occurrence
	if gErr = g.S.UpdateOccurrence(pID, oID, req.Occurrence); gErr != nil {
		log.Printf("Cannot update occurrence : %v", req.Occurrence.Name)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Cannot update Occurrences: %v", err))
	}
	return g.S.GetOccurrence(pID, oID)
}

// ListProjects returns the project id for all projects in the backing datastore.
func (g *Grafeas) ListProjects(ctx context.Context, req *prpb.ListProjectsRequest) (*prpb.ListProjectsResponse, error) {
	// TODO: support filters
	if req.PageSize == 0 {
		req.PageSize = 100
	}
	ps, nextToken, err := g.S.ListProjects(req.Filter, int(req.PageSize), req.PageToken)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to list projects")
	}
	return &prpb.ListProjectsResponse{
		Projects:      ps,
		NextPageToken: nextToken,
	}, nil
}

func (g *Grafeas) ListNotes(ctx context.Context, req *pb.ListNotesRequest) (*pb.ListNotesResponse, error) {
	pID, err := name.ParseProject(req.Parent)
	if err != nil {
		log.Printf("Error parsing name: %v", req.Parent)
		return nil, status.Error(codes.InvalidArgument, "Invalid Project name")
	}
	// TODO: support filters
	if req.PageSize == 0 {
		req.PageSize = 100
	}
	ns, nextToken, err := g.S.ListNotes(pID, req.Filter, int(req.PageSize), req.PageToken)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to list notes")
	}
	return &pb.ListNotesResponse{
		Notes:         ns,
		NextPageToken: nextToken,
	}, nil
}

func (g *Grafeas) ListOccurrences(ctx context.Context, req *pb.ListOccurrencesRequest) (*pb.ListOccurrencesResponse, error) {
	pID, err := name.ParseProject(req.Parent)
	if err != nil {
		log.Printf("Error parsing name: %v", req.Parent)
		return nil, err
	}
	// TODO: support filters - prioritizing resource url
	if req.PageSize == 0 {
		req.PageSize = 100
	}
	os, nextToken, err := g.S.ListOccurrences(pID, req.Filter, int(req.PageSize), req.PageToken)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to list occurrences")
	}
	return &pb.ListOccurrencesResponse{
		Occurrences:   os,
		NextPageToken: nextToken,
	}, nil
}

func (g *Grafeas) ListNoteOccurrences(ctx context.Context, req *pb.ListNoteOccurrencesRequest) (*pb.ListNoteOccurrencesResponse, error) {
	pID, nID, err := name.ParseNote(req.Name)
	if err != nil {
		log.Printf("Invalid note name: %v", req.Name)
		return nil, status.Error(codes.InvalidArgument, "Invalid note name")
	}
	// TODO: support filters - prioritizing resource url
	if req.PageSize == 0 {
		req.PageSize = 100
	}
	os, nextToken, gErr := g.S.ListNoteOccurrences(pID, nID, req.Filter, int(req.PageSize), req.PageToken)
	if gErr != nil {
		return nil, gErr
	}
	return &pb.ListNoteOccurrencesResponse{
		Occurrences:   os,
		NextPageToken: nextToken,
	}, nil
}

// createOccurrence validates that a note is valid and then creates an occurrence in the backing datastore.
func (g *Grafeas) createOccurrence(ctx context.Context, o *pb.Occurrence, project string) (*pb.Occurrence, error) {
	if o == nil {
		log.Print("Occurrence must not be empty.")
		return nil, status.Error(codes.InvalidArgument, "occurrence must not be empty")
	}
	npID, nID, err := name.ParseNote(o.NoteName)
	if err != nil {
		log.Printf("Invalid note name: %v", o.NoteName)
		return nil, status.Error(codes.InvalidArgument, "invalid note name")
	}
	if n, err := g.S.GetNote(npID, nID); n == nil || err != nil {
		log.Printf("Unable to getnote %v, err: %v", n, err)
		return nil, status.Errorf(codes.NotFound, "note %v not found", o.NoteName)
	}
	pID, err := name.ParseProject(project)
	if err != nil {
		log.Printf("Invalid project name: %v", project)
		return nil, status.Error(codes.InvalidArgument, "invalid project name")
	}
	if _, err := g.S.GetProject(pID); err != nil {
		log.Printf("Unable to get project %v, err: %v", pID, err)
		return nil, status.Errorf(codes.NotFound, "project %v not found", pID)
	}
	randID, err := uuid.NewRandom()
	if err != nil {
		log.Printf("Error Generating Occurrence Name: %v", err)
		return nil, status.Error(codes.Internal, "could not generate occurrence name")
	}
	o.Name = name.OccurrenceName(pID, randID.String())
	return o, g.S.CreateOccurrence(o)
}

// createNote validates that a note is valid and then creates a note in the backing datastore.
func (g *Grafeas) createNote(ctx context.Context, n *pb.Note) (*pb.Note, error) {
	if n == nil {
		log.Print("Note must not be empty.")
		return nil, status.Error(codes.InvalidArgument, "Note must not be empty")
	}
	if n.Name == "" {
		log.Printf("Note name must not be empty: %v", n.Name)
		return nil, status.Error(codes.InvalidArgument, "Note name must not be empty")
	}
	pID, _, err := name.ParseNote(n.Name)
	if err != nil {
		log.Printf("Invalid note name: %v", n.Name)
		return nil, status.Error(codes.InvalidArgument, "Invalid note name")
	}
	if _, err = g.S.GetProject(pID); err != nil {
		log.Printf("Unable to get project %v, err: %v", pID, err)
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Project %v not found", pID))
	}
	return n, g.S.CreateNote(n)
}
