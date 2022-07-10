import React from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";

import SignIn from "./views/SignIn";
import SignUp from "./views/SignUp";
import BlogOverview from "./views/BlogOverview";
import UserProfileLite from "./views/UserProfileLite";
import Errors from "./views/Errors";
import ComponentsOverview from "./views/ComponentsOverview";
import Tables from "./views/Tables";
import Tradings from "./views/Tradings";
import ProfitInfo from "./views/ProfitInfo";
import withTracker from "./withTracker";
import { DefaultLayout } from "./layouts";

import "bootstrap/dist/css/bootstrap.min.css";
//import "./shards-dashboard/styles/shards-dashboards.1.1.0.min.css";

import "./shards-dashboard/styles/scss/shards-dashboards.scss"

import { TradingBotClient } from "./service_grpc_web_pb";

//var client = new TradingBotClient("https://envoy-mnthzlygaa-de.a.run.app:443" || "http://localhost:8080", null, null);
var client = new TradingBotClient("http://localhost:8080", null, null);

//var fs = require('fs');
//var grpc = require('@grpc/grpc-js');
//const root_cert = fs.readFileSync('/etc/ssl/certs/ca-certificates.crt');
//const ssl_creds = grpc.credentials.createSsl(root_cert);
//var client = new TradingBotClient("https://envoy-mnthzlygaa-de.a.run.app:8080", ssl_creds, null);

//var client = new TradingBotClient("https://envoy-mnthzlygaa-de.a.run.app:8080", null, null);

const App = () => {
  const WithTracker = withTracker;
  console.log(new Date().toLocaleTimeString());
  return (
  <BrowserRouter>
    <Routes>
      <Route path="/" exact element={<SignIn />}/>
      <Route path="/signup" element={<SignUp />}/>
      <Route path="/blog-overview" element={
        <DefaultLayout>
          <BlogOverview client={client}/>
        </DefaultLayout>
          
      }/>
      <Route path="/user-profile-lite" element={
        <DefaultLayout>
          <UserProfileLite/>
        </DefaultLayout>
      }/>
      <Route path="/components-overview" element={
        <DefaultLayout>
          <ComponentsOverview/>
        </DefaultLayout>
      }/>
      <Route path="/tradings" element={
        <DefaultLayout>
          <Tradings client={client}/>
        </DefaultLayout>
      }/>
      <Route path="/profitinfo" element={
        <DefaultLayout>
          <ProfitInfo client={client}/>
        </DefaultLayout>
      }/>
      <Route path="/tables" element={
        <DefaultLayout>
          <Tables/>
        </DefaultLayout>
      }/>
      <Route path="/errors" element={
        <DefaultLayout>
          <Errors/>
        </DefaultLayout>
      }/>
      
      
      </Routes>
  </BrowserRouter>
  );
};

export default App;
  /*
export default () => (
  <Router basename={process.env.REACT_APP_BASENAME || ""}>
    <div>
      {routes.map((route, index) => {
        return (
          <Route
            key={index}
            path={route.path}
            exact={route.exact}
            component={withTracker(props => {
              return (
                <route.layout {...props}>
                  <route.component {...props} />
                </route.layout>
              );
            })}
          />
        );
      })}
    </div>
  </Router>
);
*/

