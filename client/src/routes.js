import React from "react";
import { Navigate } from "react-router-dom";

// Layout Types
import { DefaultLayout } from "./layouts";

import SignIn from "./views/SignIn";
import SignUp from "./views/SignUp";

// Route Views
import BlogOverview from "./views/BlogOverview";
import UserProfileLite from "./views/UserProfileLite";
import AddNewPost from "./views/AddNewPost";
import Errors from "./views/Errors";
import ComponentsOverview from "./views/ComponentsOverview";
import Tables from "./views/Tables";
import BlogPosts from "./views/BlogPosts";

export default [
  /*{
    path: "/",
    exact: true,
    layout: DefaultLayout,
    component: () => <Navigate to="/blog-overview" />
  },*/
  {
    path: "/home",
    //exact: true,
    layout: DefaultLayout,
    element: () => <Navigate to="/blog-overview" />
  },
  {
    path: "/blog-overview",
    layout: DefaultLayout,
    element: BlogOverview
  },
  {
    path: "/user-profile-lite",
    layout: DefaultLayout,
    element: UserProfileLite
  },
  {
    path: "/add-new-post",
    layout: DefaultLayout,
    element: AddNewPost
  },
  {
    path: "/errors",
    layout: DefaultLayout,
    element: Errors
  },
  {
    path: "/components-overview",
    layout: DefaultLayout,
    element: ComponentsOverview
  },
  {
    path: "/tables",
    layout: DefaultLayout,
    element: Tables
  },
  {
    path: "/blog-posts",
    layout: DefaultLayout,
    element: BlogPosts
  }
];
