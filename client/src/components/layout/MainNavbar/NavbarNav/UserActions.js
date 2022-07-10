import { getAuth, signOut } from "firebase/auth";
import React from "react";
import { Link } from "react-router-dom";
import {
  Dropdown,
  DropdownToggle,
  DropdownMenu,
  DropdownItem,
  Collapse,
  NavItem,
  NavLink
} from "shards-react";

//export var isSignOut;

export var IsSignOut = React.createContext({
  isSignOut: false,
  setIsSignOut: (auth) => {}
});

const SignOut = (value) => {
  const {setIsSignOut} = React.useContext(IsSignOut);
  setIsSignOut(true);
}

export default class UserActions extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      visible: false
    };

    this.toggleUserActions = this.toggleUserActions.bind(this);
  }

  toggleUserActions() {
    this.setState({
      visible: !this.state.visible
    });
  }

  signout() {
    SignOut(true);
    const auth = getAuth();
    signOut(auth).then(() => {
      console.log('Sign-out successful.')
    }).catch((error) => {
      console.log('An error happened.')
    });
    
  }

  render() {
    return (
      <NavItem tag={Dropdown} caret toggle={this.toggleUserActions}>
        <DropdownToggle caret tag={NavLink} className="text-nowrap px-3">
          {/*<img
            className="user-avatar rounded-circle mr-2"
            src={require("./../../../../images/avatars/0.jpg")}
            alt="User Avatar"
          />{" "}*/}
          <span className="d-none d-md-inline-block">{/*Sierra Brooks*/}</span>
        </DropdownToggle>
        <Collapse tag={DropdownMenu} right small open={this.state.visible}>
          <DropdownItem tag={(props) => <Link {...props} />} to="user-profile">
            <i className="material-icons">&#xE7FD;</i> Profile
          </DropdownItem>
          <DropdownItem tag={(props) => <Link {...props} />} to="edit-user-profile">
            <i className="material-icons">&#xE8B8;</i> Edit Profile
          </DropdownItem>
          <DropdownItem tag={(props) => <Link {...props} />} to="file-manager-list">
            <i className="material-icons">&#xE2C7;</i> Files
          </DropdownItem>
          <DropdownItem tag={(props) => <Link {...props} />} to="transaction-history">
            <i className="material-icons">&#xE896;</i> Transactions
          </DropdownItem>
          <DropdownItem divider />
          <DropdownItem tag={(props) => <Link {...props} />} to="/" className="text-danger" onClick={this.signout}>
            <i className="material-icons text-danger">&#xE879;</i> Logout
          </DropdownItem>
        </Collapse>
      </NavItem>
    );
  }
}
