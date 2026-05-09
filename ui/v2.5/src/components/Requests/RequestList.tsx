import React from "react";
import { Link } from "react-router-dom";
import { Button, Table, Badge } from "react-bootstrap";
import * as GQL from "src/core/generated-graphql";
import { LoadingIndicator } from "../Shared/LoadingIndicator";
import { ErrorMessage } from "../Shared/ErrorMessage";

const statusVariant: Record<GQL.MediaRequestStatus, string> = {
  [GQL.MediaRequestStatus.Pending]: "secondary",
  [GQL.MediaRequestStatus.Approved]: "info",
  [GQL.MediaRequestStatus.Downloading]: "primary",
  [GQL.MediaRequestStatus.Imported]: "success",
  [GQL.MediaRequestStatus.Rejected]: "warning",
  [GQL.MediaRequestStatus.Failed]: "danger",
};

export const RequestList: React.FC = () => {
  const { data, loading, error } = GQL.useFindMediaRequestsQuery({
    fetchPolicy: "cache-and-network",
  });

  if (loading && !data) return <LoadingIndicator />;
  if (error) return <ErrorMessage error={error.message} />;

  const requests = data?.findMediaRequests ?? [];

  return (
    <div className="container-fluid p-3">
      <div className="d-flex justify-content-between align-items-center mb-3">
        <h2>Media Requests</h2>
        <Link to="/requests/new">
          <Button variant="primary">New request</Button>
        </Link>
      </div>

      {requests.length === 0 ? (
        <p className="text-muted">No requests yet. Create one to get started.</p>
      ) : (
        <Table striped hover>
          <thead>
            <tr>
              <th>Title</th>
              <th>Studio</th>
              <th>Status</th>
              <th>Created</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {requests.map((r) => (
              <tr key={r.id}>
                <td>{r.title}</td>
                <td>{r.studio ?? "—"}</td>
                <td>
                  <Badge variant={statusVariant[r.status]}>{r.status}</Badge>
                </td>
                <td>{new Date(r.created_at).toLocaleString()}</td>
                <td>
                  <Link to={`/requests/${r.id}`}>
                    <Button size="sm" variant="outline-primary">
                      Open
                    </Button>
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </Table>
      )}
    </div>
  );
};
