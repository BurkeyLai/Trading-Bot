import React, {useEffect, useRef, useState} from "react";
import { 
  Container, 
  Row, 
  Col,
  Card,
  ButtonGroup, 
  Button,
  ListGroup,
  ListGroupItem,
  InputGroup,
  InputGroupAddon,
  InputGroupText,
  FormInput } from "shards-react";
import {ToastContainer} from 'react-toastify';
import PageTitle from "../components/common/PageTitle";
import UserDetails from "../components/user-profile-lite/UserDetails";
import UserAccountDetails from "../components/user-profile-lite/UserAccountDetails";
import Toast from "../utils/toastify";
import { firestoreDB } from '..';
import { doc, setDoc, getDoc } from "firebase/firestore"; 
import { getAuth, onAuthStateChanged } from "firebase/auth";
import useEnhancedEffect from "@mui/material/utils/useEnhancedEffect";

var userRef = null;
var userSnap = null;
const UserProfileLite = () => {
  var userDetails = UserDetails.defaultProps;

  const [updateUserDatil, setUpdateUserDatil] = useState(false);
  const [newHuobiKey, setNewHuobiKey] = useState(false);
  const [newBinanceKey, setNewBinanceKey] = useState(false);

  const [huobiAPIKey, updateHuobiAPIKey] = useState('');
  const [huobiSecretKey, updateHuobiSecretKey] = useState('');
  const [binanceSpotAPIKey, updateBinanceSpotAPIKey] = useState('');
  const [binanceFutureAPIKey, updateBinanceFutureAPIKey] = useState('');
  const [binanceSpotSecretKey, updateBinanceSpotSecretKey] = useState('');
  const [binanceFutureSecretKey, updateBinanceFutureSecretKey] = useState('');

  const onClickNewHuobiKey = () => {
    if (!newHuobiKey) {
      setNewHuobiKey(true)
      
      if (userSnap.exists) {
        if (userSnap.get('huobi_apikey')) {
          updateHuobiAPIKey(userSnap.get('huobi_apikey'))
        }
        if (userSnap.get('huobi_secretkey')) {
          updateHuobiSecretKey(userSnap.get('huobi_secretkey'))
        }
      }
    } else {
      setNewHuobiKey(false)
    }
  };
  const onClickNewBinanceKey = () => {
    if (!newBinanceKey) {
      setNewBinanceKey(true)

      if (userSnap.exists) {
        if (userSnap.get('binance_spot_apikey')) {
          updateBinanceSpotAPIKey(userSnap.get('binance_spot_apikey'))
        }
        if (userSnap.get('binance_spot_secretkey')) {
          updateBinanceSpotSecretKey(userSnap.get('binance_spot_secretkey'))
        }
        if (userSnap.get('binance_future_apikey')) {
          updateBinanceFutureAPIKey(userSnap.get('binance_future_apikey'))
        }
        if (userSnap.get('binance_future_secretkey')) {
          updateBinanceFutureSecretKey(userSnap.get('binance_future_secretkey'))
        }
      }
    } else {
      setNewBinanceKey(false)
    }
  };
  const confirmNewHuobiKey = () => {
    setNewHuobiKey(false)
    if (userRef != null) {
      setDoc(userRef, { huobi_apikey: huobiAPIKey, huobi_secretkey: huobiSecretKey}, { merge: true });
    }
    Toast('success', true, '儲存成功');
  }
  const confirmNewBinanceKey = () => {
    setNewBinanceKey(false)
    if (userRef != null) {
      setDoc(userRef, { binance_spot_apikey: binanceSpotAPIKey, binance_spot_secretkey: binanceSpotSecretKey,
        binance_future_apikey: binanceFutureAPIKey, binance_future_secretkey: binanceFutureSecretKey}, { merge: true });
    }
    Toast('success', true, '儲存成功');
  }

  useEffect(() => {
    const auth = getAuth();
    onAuthStateChanged(auth, async (f_user) => {
      if (f_user) {
        userRef = doc(firestoreDB, 'User', f_user.uid);
        userSnap = await getDoc(userRef);
        if (userSnap.exists) {
          userDetails.name = userSnap.get('name');
          userDetails.jobTitle = userSnap.get('email');
          setUpdateUserDatil(true);
        }
      } else {
        // User is signed out
      }
    });
  }, []);

  return (
  <Container fluid className="main-content-container px-4">
    <ToastContainer />
    <Row noGutters className="page-header py-4">
      <PageTitle title="使用者檔案" /*subtitle="Overview"*/ md="12" className="ml-sm-auto mr-sm-auto" />
    </Row>
    {updateUserDatil ?
    <Row>
      <Col lg="4">
        <UserDetails userDetails={userDetails}/>
      </Col>
      {/*
      <Col lg="8">
        <UserAccountDetails />
      </Col>
      */}
    </Row> : null}
    <strong className="text-muted d-block my-2">
        API 授權
    </strong>
    <Row>
      <Col>
        <ButtonGroup className="mb-3 mr3">
          <Button onClick={onClickNewHuobiKey} outline theme="success">
            火幣
          </Button>
        </ButtonGroup>
      </Col>
    </Row>
    {newHuobiKey ?
      (
        <Row>
          <Col lg="6" className="mb-4">
            <Card small>
              <ListGroup flush>
                <ListGroupItem className="px-3">
                  <Row form>
                    <Col md="6" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        API Key
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormInput value={huobiAPIKey} onChange={evt => updateHuobiAPIKey(evt.target.value)}/>
                        </InputGroup>
                      </div>
                    </Col>
                    <Col md="6" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        Secret Key
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormInput value={huobiSecretKey} onChange={evt => updateHuobiSecretKey(evt.target.value)}/>
                        </InputGroup>
                      </div>
                    </Col>
                  </Row>
                  <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                    <Button onClick={confirmNewHuobiKey} outline theme="success" >
                      儲存
                    </Button>
                  </div>
                </ListGroupItem>
              </ListGroup>
            </Card>
          </Col>
        </Row>
      )
      : null}
    <Row>
      <Col>
        <ButtonGroup className="mb-3 mr3">
          <Button onClick={onClickNewBinanceKey}outline theme="success">
            幣安
          </Button>
        </ButtonGroup>
      </Col>
    </Row>
    {newBinanceKey ?
      (
        <Row>
          <Col lg="12" className="mb-4">
            <Card small>
              <ListGroup flush>
                <ListGroupItem className="px-3">
                  <Row form>
                    <Col md="6" className="form-group">
                      <strong className="d-block mb-2">
                        現貨
                      </strong>
                    </Col>
                    <Col md="6" className="form-group">
                      <strong className="d-block mb-2">
                        合約
                      </strong>
                    </Col>
                  </Row>
                  <Row form>
                    <Col md="3" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        API Key
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormInput value={binanceSpotAPIKey} onChange={evt => updateBinanceSpotAPIKey(evt.target.value)}/>
                        </InputGroup>
                      </div>
                    </Col>
                    <Col md="3" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        Secret Key
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormInput value={binanceSpotSecretKey} onChange={evt => updateBinanceSpotSecretKey(evt.target.value)}/>
                        </InputGroup>
                      </div>
                    </Col>
                    <Col md="3" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        API Key
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormInput value={binanceFutureAPIKey} onChange={evt => updateBinanceFutureAPIKey(evt.target.value)}/>
                        </InputGroup>
                      </div>
                    </Col>
                    <Col md="3" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        Secret Key
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormInput value={binanceFutureSecretKey} onChange={evt => updateBinanceFutureSecretKey(evt.target.value)}/>
                        </InputGroup>
                      </div>
                    </Col>
                  </Row>
                  <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                    <Button onClick={confirmNewBinanceKey} outline theme="success" >
                      儲存
                    </Button>
                  </div>
                </ListGroupItem>
              </ListGroup>
            </Card>
          </Col>
        </Row>
      )
      : null}
  </Container>
);}

export default UserProfileLite;
