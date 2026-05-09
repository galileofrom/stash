import React from "react";
import { useParams, Link } from "react-router-dom";
import { Button, Table, Badge, Card } from "react-bootstrap";
import * as GQL from "src/core/generated-graphql";
import { LoadingIndicator } from "../Shared/LoadingIndicator";
import { ErrorMessage } from "../Shared/ErrorMessage";

function formatBytes(n: number) {
  if (!n) return "—";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let i = 0;
  let v = n;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i += 1;
  }
  return `${v.toFixed(1)} ${units[i]}`;
}

export const RequestDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { data, loading, error, refetch } = GQL.useFindMediaRequestQuery({
    variables: { id },
    fetchPolicy: "cache-and-network",
  });

  const [search, { loading: searching }] = GQL.useMediaRequestSearchMutation();
  const [approve, { loading: approving }] = GQL.useMediaRequestApproveMutation();
  const [reject, { loading: rejecting }] = GQL.useMediaRequestRejectMutation();

  if (loading && !data) return <LoadingIndicator />;
  if (error) return <ErrorMessage error={error.message} />;
  const req = data?.findMediaRequest;
  if (!req) return <p className="p-3 text-muted">Request not found.</p>;

  return (
    <div className="container-fluid p-3">
      <div className="d-flex justify-content-between align-items-start mb-3">
        <div>
          <h2 className="mb-1">{req.title}</h2>
          {req.studio && <p className="text-muted mb-2">{req.studio}</p>}
          <Badge variant="secondary">{req.status}</Badge>
        </div>
        <div className="d-flex gap-2">
          <Button
            variant="outline-primary"
            disabled={searching}
            onClick={async () => {
              await search({ variables: { id } });
              refetch();
            }}
          >
            {searching ? "Searching…" : "Search Prowlarr"}
          </Button>
          <Button
            variant="outline-warning"
            disabled={rejecting}
            onClick={async () => {
              await reject({ variables: { id } });
              refetch();
            }}
          >
            Reject
          </Button>
          <Link to="/requests">
            <Button variant="link">← Back</Button>
          </Link>
        </div>
      </div>

      {req.notes && (
        <Card className="mb-3">
          <Card.Body className="py-2">
            <small className="text-muted">{req.notes}</small>
          </Card.Body>
        </Card>
      )}

      <h4>Releases</h4>
      {req.releases.length === 0 ? (
        <p className="text-muted">
          No releases yet. Click <em>Search Prowlarr</em> to query indexers.
        </p>
      ) : (
        <Table striped hover size="sm">
          <thead>
            <tr>
              <th>Indexer</th>
              <th>Title</th>
              <th>Protocol</th>
              <th>Size</th>
              <th>Seeders</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {req.releases.map((r) => (
              <tr key={r.id}>
                <td>{r.indexer}</td>
                <td>{r.title}</td>
                <td>
                  <Badge variant={r.protocol === GQL.MediaReleaseProtocol.Usenet ? "info" : "secondary"}>
                    {r.protocol.toLowerCase()}
                  </Badge>
                </td>
                <td>{formatBytes(r.size)}</td>
                <td>{r.protocol === GQL.MediaReleaseProtocol.Torrent ? r.seeders : "—"}</td>
                <td>
                  <Button
                    size="sm"
                    variant="success"
                    disabled={approving}
                    onClick={async () => {
                      await approve({ variables: { release_id: r.id } });
                      refetch();
                    }}
                  >
                    Approve
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </Table>
      )}
    </div>
  );
};
