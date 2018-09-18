// @flow
import React from 'react';
import { TextField, Button } from '@material-ui/core';
import _ from 'lodash';
import PropTypes from 'prop-types'

import w2mService from '../services/w2mService';

// TODO add refresh? Yeah sure, so they don't have to log the hell back in

class Home extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      name: "",
      password: "",

      userId: null,
      userAvailability: [],
    };
  }

  handleLogin() {
    const { id, name, password } = this.state;
    w2mService.login(id, name, password)
    .then(function(userId) {
      this.setState({
        userId,
        // TODO populate input availability based on the logged-in user,
        // i.e. pull out this user's availability and plop it in
        userAvailability: _.find(this.props.availability, x => x.id === userId),
      });
    });
  }

  // TODO should be automatic?
  handleSave() {
    const { userId, userAvailability } = this.state;
    const { id } = this.props;
    w2mService.saveTimes(userId, id, userAvailability)
    .then(function() {
      //- TODO: WHAT DOES THIS RETURN???
      // this.setState(??????????????);
      // maybe alert? lel
    });
  }

  render() {
    // TODO convert to UI
    const availabilityView = JSON.encode(this.props.availability);

    let content;
    if (!this.state.userId) {
      content = (
        <div className="login">
          <TextField label="Name" value={this.state.name} onChange={this.handleChange('name')}/>
          <TextField label="Password" value={this.state.password} onChange={this.handleChange('password')}/>
          <Button onClick={this.handleLogin}>Login</Button>
        </div>
      );
    } else {
      content = (
        <div>
          User's availability goes here
          <Button onClick={this.handleSave}>Save</Button>
        </div>
      );
    }

    return (
      <div>
        {availabilityView}
        {content}
      </div>
    );
  }
}

// TODO set proptypes
// id, code, timeZone, availability
Home.propTypes = {
  id: PropTypes.string.required,
  code: PropTypes.string.required,
  timeZone: PropTypes.string,
  availability: PropTypes.object.required,
};

export default Home;
