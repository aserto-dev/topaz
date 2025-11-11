#!/usr/bin/env bash

jq '{
  assertions: [
    .assertions[] |
    {
      evaluation: {
        subject: {
          type: .check.subject_type,
          id: .check.subject_id
        },
        action: {
          name: .check.relation
        },
        resource: {
          type: .check.object_type,
          id: .check.object_id
        }
      },
      expected: .expected
    }
  ]
}' $1 
