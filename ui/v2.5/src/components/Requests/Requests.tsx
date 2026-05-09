import React from "react";
import { Route, Switch } from "react-router-dom";
import { Helmet } from "react-helmet";
import { RequestList } from "./RequestList";
import { RequestCreate } from "./RequestCreate";
import { RequestDetail } from "./RequestDetail";

const Requests: React.FC = () => {
  return (
    <>
      <Helmet>
        <title>Requests</title>
      </Helmet>
      <Switch>
        <Route exact path="/requests" component={RequestList} />
        <Route exact path="/requests/new" component={RequestCreate} />
        <Route path="/requests/:id" component={RequestDetail} />
      </Switch>
    </>
  );
};

export default Requests;
