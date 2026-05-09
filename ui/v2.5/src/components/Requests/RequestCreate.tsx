import React, { useState } from "react";
import { useHistory } from "react-router-dom";
import { Form, Button } from "react-bootstrap";
import * as GQL from "src/core/generated-graphql";

export const RequestCreate: React.FC = () => {
  const history = useHistory();
  const [title, setTitle] = useState("");
  const [studio, setStudio] = useState("");
  const [notes, setNotes] = useState("");

  const [createRequest, { loading, error }] = GQL.useMediaRequestCreateMutation({
    refetchQueries: [GQL.FindMediaRequestsDocument],
  });

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!title.trim()) return;

    const result = await createRequest({
      variables: {
        input: {
          title: title.trim(),
          studio: studio.trim() || undefined,
          notes: notes.trim() || undefined,
        },
      },
    });

    const id = result.data?.mediaRequestCreate.id;
    if (id) history.push(`/requests/${id}`);
  }

  return (
    <div className="container p-3" style={{ maxWidth: 640 }}>
      <h2>New request</h2>
      <Form onSubmit={onSubmit}>
        <Form.Group className="mb-3">
          <Form.Label>Title</Form.Label>
          <Form.Control
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Scene or video title"
            required
          />
        </Form.Group>

        <Form.Group className="mb-3">
          <Form.Label>Studio</Form.Label>
          <Form.Control
            value={studio}
            onChange={(e) => setStudio(e.target.value)}
            placeholder="Optional studio name"
          />
        </Form.Group>

        <Form.Group className="mb-3">
          <Form.Label>Notes</Form.Label>
          <Form.Control
            as="textarea"
            rows={3}
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
          />
        </Form.Group>

        {error && <p className="text-danger">{error.message}</p>}

        <Button type="submit" variant="primary" disabled={loading || !title}>
          {loading ? "Creating…" : "Create"}
        </Button>
      </Form>
    </div>
  );
};
