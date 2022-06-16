import React, {useEffect, useState} from "react";
import PropTypes from "prop-types";
import {
  Card,
  CardHeader,
  Button,
  ListGroup,
  ListGroupItem,
  Progress
} from "shards-react";
import { getAuth, onAuthStateChanged } from "firebase/auth";
import { doc, getDoc} from "firebase/firestore";
import { firestoreDB } from '../..';

var userRef = null;
var userSnap = null;

//const UserDetails = ({ userDetails }) => (
const UserDetails = ({ userDetails }) => {
  
  /*
  const [userName, setUserName] = useState(userDetails.name);
  const [userEmail, setUserEmail] = useState(userDetails.jobTitle);

  useEffect(() => {
    setUserName(userDetails.name);
  }, [userDetails.name]);

  useEffect(() => {
    setUserEmail(userDetails.jobTitle);
  }, [userDetails.jobTitle]);

  
  useEffect(() => {
    const auth = getAuth();
      onAuthStateChanged(auth, async (f_user) => {
        
        if (f_user) {
          userRef = doc(firestoreDB, 'User', f_user.uid);
          userSnap = await getDoc(userRef);
          if (userSnap.exists) {
            
            userDetails.name = userSnap.get('name');
            userDetails.jobTitle = userSnap.get('email');
            console.log(userDetails.name)
            console.log(userDetails.jobTitle)
          }
        } 
      });
  }, []);
  */
  
  return (
  <Card small className="mb-4 pt-3">
    <CardHeader className="border-bottom text-center">
      <div className="mb-3 mx-auto">
        {/*<img
          className="rounded-circle"
          src={userDetails.avatar}
          alt={userDetails.name}
          width="110"
        />*/}
      </div>
      <h4 className="mb-2">{userDetails.name}</h4>
      <span className="text-muted d-block mb-2">{userDetails.jobTitle}</span>
      
      {/*
      <Button pill outline size="sm" className="mb-2">
        <i className="material-icons mr-1">person_add</i> Follow
      </Button>
      */}
    </CardHeader>
    <ListGroup flush>
      <ListGroupItem className="px-4">
        {/*
        <div className="progress-wrapper">
          <strong className="text-muted d-block mb-2">
            {userDetails.performanceReportTitle}
          </strong>
          <Progress
            className="progress-sm"
            value={userDetails.performanceReportValue}
          >
            <span className="progress-value">
              {userDetails.performanceReportValue}%
            </span>
          </Progress>
        </div>
        */}
      </ListGroupItem>
      <ListGroupItem className="p-4">
        {/*
        <strong className="text-muted d-block mb-2">
          {userDetails.metaTitle}
        </strong>
        <span>{userDetails.metaValue}</span>
        */}
      </ListGroupItem>
    </ListGroup>
  </Card>
//);
  );}

UserDetails.propTypes = {
  /**
   * The user details object.
   */
  userDetails: PropTypes.object
};

UserDetails.defaultProps = {
  userDetails: {
    //name: "Sierra Brooks",
    name: "",
    avatar: require("./../../images/avatars/0.jpg"),
    //jobTitle: "Project Manager",
    jobTitle: "",
    performanceReportTitle: "Workload",
    performanceReportValue: 74,
    metaTitle: "Description",
    metaValue:
      "Lorem ipsum dolor sit amet consectetur adipisicing elit. Odio eaque, quidem, commodi soluta qui quae minima obcaecati quod dolorum sint alias, possimus illum assumenda eligendi cumque?"
  }
};

export default UserDetails;
