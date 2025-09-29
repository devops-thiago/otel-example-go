package models

import (
    "testing"
    "time"
)

func TestToResponse(t *testing.T) {
    now := time.Now()
    u := &User{ID: 7, Name: "N", Email: "e@x", Bio: "b", CreatedAt: now, UpdatedAt: now}
    r := u.ToResponse()
    if r.ID != 7 || r.Name != "N" || r.Email != "e@x" || r.Bio != "b" {
        t.Fatalf("unexpected response: %+v", r)
    }
}


