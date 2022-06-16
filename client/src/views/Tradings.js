import React, { useEffect, useState } from "react";
import { 
  Container, 
  Row, 
  Col, 
  Card, 
  CardHeader, 
  CardBody, 
  ButtonGroup, 
  Button,
  ListGroup,
  ListGroupItem,
  InputGroup,
  FormSelect,
  InputGroupAddon,
  InputGroupText,
  FormInput, 
  symbol} from "shards-react";
//import { v4 as uuidv4 } from 'uuid';
import { getAuth, onAuthStateChanged } from "firebase/auth";
import { firestoreDB } from '..';
import { doc, setDoc, getDoc, getDocs, query, orderBy, limit, where, collection, onSnapshot, collectionGroup } from "firebase/firestore"; 
import PageTitle from "../components/common/PageTitle";
import {ToastContainer} from 'react-toastify';
import Toast from "../utils/toastify";
import { User, 
  Connect, 
  Message, 
  Exchange, 
  ExchangeConfig, 
  DepositAddresses, 
  MarketInfoRequest,
  AccountBalanceRequest,
  CreateOrderRequest, } from "../service_pb";


const exchNum = 2, exch1 = "火幣", exch2 = "幣安", conNum = 2, type1 = "U本位", type2 = "幣本位";
const exchNameArray = ['huobi', 'binance'];
const modeNameArray = ['spot', 'future'];
const exchcfg = new ExchangeConfig();
const user = new User();
const connect = new Connect();
const msg = new Message();
var userRef = null;
var userSnap = null;

const Tradings = ({ client }) => {
    
    //const symbolArray1 = [], symbolArray2 = [];
    //const symbolTest = ['1', '2', '3'];
    const [briefBotInfoArray, setBriefBotInfoArray] = useState([]);
    //const [updateBriefBotInfoArray, setUpdateBriefBotInfoArray] = useState(false);

    const [symbolArray1, setSymbolArray1] = useState([]);
    const [symbolArray2, setSymbolArray2] = useState([]);
    const [selectedExch, setSelectedExch] = useState(exch1);
    const [selectedType, setSelectedType] = useState(type1);
    const [selectedSymbol, setSelectedSymbol] = useState('');
    //const [selectedSymbol2, setSelectedSymbol2] = useState('');
    const [selectedLeverage, setSelectedLeverage] = useState('1');
    const [maxDrawdown, setMaxDrawdown] = useState('');
    const [maxAmount, setMaxAmount] = useState(0);
    const [future, setFuture] = useState(false);
    const [newTrading, setNewTrading] = useState(false);
    const [conservative, setConservative] = useState(true);
    const [symbolBalance, setSymbolBalance] = useState('');
    
    const askMarketSymbols = () => {
      
      if (user.getId() !== '') {
        const req = new MarketInfoRequest();
        msg.setContent("Ask for market info...");
        msg.setTimestamp(new Date().toLocaleTimeString());
        req.setMsg(msg);
        (future) ? req.setMode('future') : req.setMode('spot');
        (selectedExch === '火幣') ? req.setExchange('Huobi') : req.setExchange('Binance');
        (selectedType === 'U本位') ? req.setType('usdt') : req.setType('coin');
        //console.log(req)
        client.marketInfo(req, null, (err, resp) => {
          if (resp) {
            setSymbolArray1(['請選擇...', ...resp.getSymbolsList()]);
          } else {
            setSymbolArray1(['請選擇...']);
            Toast('error', true, "API 發生錯誤");
            //console.log(err)
          }
        });
      }
    }
    const askAccountBalance = () => {
      if (user.getId() !== '') {
        const req = new AccountBalanceRequest();
        msg.setContent("Ask for account balance...");
        msg.setTimestamp(new Date().toLocaleTimeString());
        req.setMsg(msg);
        (future) ? req.setMode('future') : req.setMode('spot');
        (selectedExch === '火幣') ? req.setExchange('Huobi') : req.setExchange('Binance');
        req.setSymbol('usdt');
        client.accountBalance(req, null, (err, resp) => {
          if (resp) {
            setSymbolBalance(resp.getBalance());
          }
        });
      }
    }
    
    const launchCreateOrder = () => {
      if (user.getId() !== '') {
        const req = new CreateOrderRequest();
        msg.setContent("Ask for account balance...");
        msg.setTimestamp(new Date().toLocaleTimeString());
        req.setMsg(msg);
        (future) ? req.setMode('future') : req.setMode('spot');
        (conservative) ? req.setDroppercent('0.03') : req.setDroppercent('0.01');
        (conservative) ? req.setGouppercent('0.05') : req.setGouppercent('0.01');
        (selectedExch === '火幣') ? req.setExchange('Huobi') : req.setExchange('Binance');
        (selectedType === 'U本位') ? req.setType('usdt') : req.setType('coin');
        req.setSymbol(selectedSymbol);
        req.setCycletype('');
        req.setLeverage(selectedLeverage);
        req.setMaxdrawdown(maxDrawdown);
        req.setWithdrawspot('');
        req.setQuantity(maxAmount.toString());
        client.createOrder(req, null, (err, resp) => {
          console.log(resp.getContent())
          //console.log(resp.getOrderid())
          if (resp.getBotactive()) {
            if (userRef != null) { 
              const storeBots = async () => {
                userSnap = await getDoc(userRef);
                if (userSnap.exists) {
                  const strs = req.getSymbol().split('/');
                  const obj = {
                    bot_active : resp.getBotactive(),
                    mode : req.getMode(),
                    drop_percent : req.getDroppercent(),
                    go_up_percent : req.getGouppercent(),
                    exchange : req.getExchange().toLowerCase(),
                    type : req.getType(),
                    symbol : strs[0] + strs[1],
                    cycle_type : req.getCycletype(),
                    leverage : req.getLeverage(),
                    max_drawdown : req.getMaxdrawdown(),
                    withdraw_spot : req.getWithdrawspot(),
                    quantity : req.getQuantity(),
                    average_price: '',
                    order_id_list: []
                  };
                  if (userSnap.get("bots_array")) {
                    var arr = userSnap.get("bots_array");
                    arr = [...arr, obj]
                    setDoc(userRef, {
                      'bots_array': arr
                    }, { merge: true });
                  } else {
                    setDoc(userRef, {
                      'bots_array': [obj]
                    }, { merge: true });
                  }
                }
                /*
              setDoc(userRef, 
              { [req.getExchange().toLowerCase()] : 
                { [req.getMode()] : 
                  { [strs[0] + strs[1]] : 
                    {
                      bot_active : resp.getBotactive(),
                      mode : req.getMode(),
                      drop_percent : req.getDroppercent(),
                      go_up_percent : req.getGouppercent(),
                      exchange : req.getExchange().toLowerCase(),
                      type : req.getType(),
                      symbol : req.getSymbol(),
                      cycle_type : req.getCycletype(),
                      leverage : req.getLeverage(),
                      max_drawdown : req.getMaxdrawdown(),
                      withdraw_spot : req.getWithdrawspot(),
                      quantity : req.getQuantity(),
                      average_price: '',
                      order_id_list: []
                    } 
                  } 
                } 
              }, { merge: true });
              */
                return '儲存成功'
              }

              storeBots().then((res) => {
                console.log(res);
              });
            }
          }
        });
      }
    }

    const tradingMethod1 = () => {
      setFuture(false)
    };
    const tradingMethod2 = () => {
      setFuture(true)
    };
    const addNewTrading = () => {
      if (!newTrading) {
        setNewTrading(true);
        askMarketSymbols();
        askAccountBalance();
      } else {
        setNewTrading(false);
      }
    };
    const strategy1 = () => {
      setConservative(true)
    };
    const strategy2 = () => {
      setConservative(false)
    };

    const confirmNewTrading = () => {
      //console.log(selectedSymbol1)
      //console.log(selectedSymbol2)
      if (selectedSymbol === '') {
        Toast('error', true, '請至少選擇一幣種');
      } else if (maxDrawdown === '') {
        Toast('error', true, '請輸入最大回撤');
      } else if (maxAmount === '') {
        Toast('error', true, '請輸入首單金額');
      } else {
        setNewTrading(false)
        launchCreateOrder();
      }
      
      //const auth = getAuth();
      //onAuthStateChanged(auth, (f_user) => {
      //  if (f_user) {
      //    user.setId(f_user.uid);
      //    user.setName(f_user.displayName);
      //    
      //    const msg = new Message();
      //    msg.setUser(user);
      //    msg.setContent("Ask for market info...");
      //    msg.setTimestamp(new Date().toLocaleTimeString());
      //    client.marketInfo(msg, null, () => {
      //    });
      //  } else {
      //    // User is signed out
      //  }
      //});
    };

    const updateBriefBotInfo = (info) => {
      const auth = getAuth();
      onAuthStateChanged(auth, async (f_user) => {
        if (f_user) {
          /*
          userRef = doc(firestoreDB, 'User', f_user.uid);
          userSnap = await getDoc(userRef);
          if (userSnap.exists) {
            if (userSnap.get("bots_array")) {
              var arr = userSnap.get("bots_array");
              for (var i = 0; i < arr.length; i++) {
                console.log(arr[i].symbol)
              }
              //const obj = {
              //  exchange: 'binance',
              //  symbol: 'BTCUSDT',
              //  mode: 'spot'
              //}
              //arr = [...arr, obj]
              //setDoc(userRef, {
              //  'bots_array': arr
              //}, { merge: true });
            } else {
              const obj = {
                exchange: 'binance',
                symbol: 'ETHUSDT',
                mode: 'spot'
              }
              setDoc(userRef, {
                'bots_array': [obj]
              }, { merge: true });
            }
          }
          */

          userSnap = await getDoc(userRef);
          if (userSnap.exists) {
            if (userSnap.get("bots_array")) {
              var arr = userSnap.get("bots_array");
              var updateArray = [];
              for (var i = 0; i < arr.length; ++i) {
                const tbl = 
                {
                  brfIdx: i,
                  brfExch: (arr[i].exchange == 'huobi') ? exch1 : exch2,
                  brfSymbol: arr[i].symbol,
                  brfAvgPrice: arr[i].average_price,
                  brfMode: (arr[i].mode == 'spot') ? '現貨' : '合約',
                  brfQty: arr[i].quantity
                };
                updateArray = [...updateArray, tbl];
              }
              setBriefBotInfoArray(updateArray);
            }
          }
        }
      });
    }

    const updateBotInfo = (info) => {
      if (info.getModelname() != "") {
        if (userRef != null) { 
          const updateBots = async () => {
            userSnap = await getDoc(userRef);
            if (userSnap.exists) {
              const botExchname = info.getExch();
              const botMode = info.getMode();
              const botSymbol = info.getModelname();
              if (userSnap.get("bots_array")) {
                var arr = userSnap.get("bots_array");
                var i = 0, found = false;
                for (i = 0; i < arr.length; i++) {
                  if (arr[i].exchange == botExchname &&
                      arr[i].mode == botMode &&
                      arr[i].symbol == botSymbol) {

                    found = true;
                    break;
                  }
                }
                if (found) {
                  arr[i].average_price = info.getAvgprice();
                  arr[i].order_id_list = info.getOrderidlistList();
                  setDoc(userRef, {
                    'bots_array': arr
                  }, { merge: true });
                }
              }
            }
            return '更新成功'
          }

          updateBots().then((res) => {
            console.log(res);
          });
        }
        /*
        setDoc(userRef, 
        { [info.getExch()] :
          { [info.getMode()] :
            { [info.getModelname()] : 
              {
                average_price: info.getAvgprice(),
                order_id_list: info.getOrderidlistList()
              } 
            }
          }
        }, { merge: true });
        */
        //setUpdateBriefBotInfoArray(true);
        //setUpdateBriefBotInfoArray(false);
        updateBriefBotInfo(info);
      }
    }

    const askCreateStream = () => {
      const auth = getAuth();
      onAuthStateChanged(auth, async (f_user) => {
        
        if (f_user) {
          userRef = doc(firestoreDB, 'User', f_user.uid);
          userSnap = await getDoc(userRef);

          var exchs = [];
          for (var i = 0; i < exchNum; i++) {
            const exch = new Exchange();
            const n = '', spot_pub_k = '', spot_sec_k = '', future_pub_k = '', future_sec_k = '', addr = new DepositAddresses();
            if (i === 0) {
              n = 'Huobi';
              addr.setAddrnum('2');
              addr.setBtcaddr('');
              addr.setEthaddr('');
              if (userSnap.exists) {
                spot_pub_k = (userSnap.get('huobi_apikey')) ? userSnap.get('huobi_apikey') : '';
                spot_sec_k = (userSnap.get('huobi_secretkey')) ? userSnap.get('huobi_secretkey') : '';
              }
            } else if (i === 1) {
              n = 'Binance';
              addr.setAddrnum('2');
              addr.setBtcaddr('');
              addr.setEthaddr('');
              if (userSnap.exists) {
                spot_pub_k = (userSnap.get('binance_spot_apikey')) ? userSnap.get('binance_spot_apikey') : '';
                spot_sec_k = (userSnap.get('binance_spot_secretkey')) ? userSnap.get('binance_spot_secretkey') : '';
                future_pub_k = (userSnap.get('binance_future_apikey')) ? userSnap.get('binance_future_apikey') : '';
                future_sec_k = (userSnap.get('binance_future_secretkey')) ? userSnap.get('binance_future_secretkey') : '';
              }
            }
            //exchs[i] = {exchname: n, publickey: pub_k, secretkey: sec_k, depoaddr: addr};
            exch.setExchname(n);
            exch.setSpotpublickey(spot_pub_k);
            exch.setSpotsecretkey(spot_sec_k);
            exch.setFuturepublickey(future_pub_k);
            exch.setFuturesecretkey(future_sec_k);
            exch.setDepoaddr(addr);
            exchs.push(exch);
          }

          exchcfg.setExchsList(exchs);
          user.setId(f_user.uid);
          user.setName(f_user.displayName);
          user.setExchcfg(exchcfg);
          msg.setUser(user);
          connect.setUser(user);
          connect.setActive(true);
          
          var msgStream = client.createStream(connect, {});
          msgStream.on("data", (response) => {
            const s_user = response.getUser();
            const id = s_user.getId();
            const username = s_user.getName();
            const messageContent = response.getContent();
            const timestamp = response.getTimestamp();
            console.log(id + "[" + username + "]: " + messageContent);
            if (response.getBotinfo()) {
              updateBotInfo(response.getBotinfo());
            }
          });
        } else {
          // User is signed out
        }
      });
    }
    
    //useEffect(() => {
    //  if (updateBriefBotInfoArray) {
    //    const auth = getAuth();
    //    onAuthStateChanged(auth, async (f_user) => {
    //      if (f_user) {
    //        userRef = doc(firestoreDB, 'User', f_user.uid);
    //        userSnap = await getDoc(userRef);
    //        if (userSnap.exists) {
    //          //spot_pub_k = (userSnap.get('huobi_apikey')) ? userSnap.get('huobi_apikey') : '';
    //          //spot_sec_k = (userSnap.get('huobi_secretkey')) ? userSnap.get('huobi_secretkey') : '';
    //        }
    //      }
    //    });
    //  }
    //}, [updateBriefBotInfoArray])

    useEffect(() => {
      askCreateStream();
      updateBriefBotInfo();

      //const auth = getAuth();
      //onAuthStateChanged(auth, async (f_user) => {
      //  userRef = doc(firestoreDB, 'User', f_user.uid);
      //  //const q = query(collection(firestoreDB, "User"), where("binance", "==", "spot"));
      //  const qq = query(collection(firestoreDB, "User"), where("bot_active", "==", true));
      //  //console.log(q)
      //  //console.log(qq)
      ////  if (f_user) {
      //  const unsub = onSnapshot(qq, (querySnapshot) => {
      //    querySnapshot.forEach((doc) => {
      //      console.log(doc);
      //    });
      ////      console.log(doc);
      ////    console.log("Current data: ", doc.data());
      //  });
      ////  }
      //});
    }, [])

    useEffect(() => {
      askMarketSymbols();
      askAccountBalance();
    }, [selectedExch])

    useEffect(() => {
      //askExchKeyConfig();
      //setNewTrading(false);
      askMarketSymbols();
      askAccountBalance();
      if (future) {
        setSelectedLeverage('20');
      } else {
        setSelectedLeverage('1');
      }
      
    }, [future])

    //useEffect(() => {
    //  setSelectedSymbol2Ratio(100 - selectedSymbol1Ratio);
    //}, [selectedSymbol1Ratio])

    //useEffect(() => {
    //  setSelectedSymbol1Ratio(100 - selectedSymbol2Ratio);
    //}, [selectedSymbol2Ratio])

    return (
    <Container fluid className="main-content-container px-4">
    {/* Page Header */}
    <ToastContainer />
    <Row noGutters className="page-header py-4">
      <PageTitle sm="4" title="交易" /*subtitle="Blog Posts"*/ className="text-sm-left" />
    </Row>

    <strong className="text-muted d-block my-2">
        交易方式
    </strong>
    <Row>
      <Col>
        <ButtonGroup className="mb-3 mr-3">
          <Button onClick={tradingMethod1} outline={future} theme="success">
            現貨
          </Button>
          <Button onClick={tradingMethod2} outline={!future} theme="success">
            合約
          </Button>
        </ButtonGroup>
        <ButtonGroup className="mb-3 mr3">
          <Button onClick={addNewTrading} outline theme="success">
            新增交易對
          </Button>
        </ButtonGroup>
      </Col>
    </Row>

    {newTrading ? 
      (
        <Row>
          <Col lg="6" className="mb-4">
            <Card small>
              <CardHeader className="border-bottom">
                <h6 className="m-0">策略開通</h6>
              </CardHeader>

              <ListGroup flush>
                <ListGroupItem className="px-3">
                  <Row form>
                    <Col className="form-group">
                      <strong className="text-muted d-block mb-2">
                        策略選擇
                      </strong>
                      <div>
                        <ButtonGroup className="mb-2">
                          <Button onClick={strategy1} outline={!conservative} theme="success">
                            保守
                          </Button>
                          <Button onClick={strategy2} outline={conservative} theme="success">
                            激進
                          </Button>
                        </ButtonGroup>
                      </div>
                    </Col>
                  </Row>
                  
                  <Row form>
                    <Col md="6" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        交易所
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormSelect value={selectedExch} onChange={evt => setSelectedExch(evt.target.value)}>
                            <option>{exch1}</option>
                            <option>{exch2}</option>
                          </FormSelect>
                        </InputGroup>
                      </div>
                    </Col>
                    <Col md="6" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        本位類型
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormSelect value={selectedType} onChange={evt => setSelectedType(evt.target.value)}>
                            <option>{type1}</option>
                            <option>{type2}</option>
                          </FormSelect>
                        </InputGroup>
                      </div>
                    </Col>
                  </Row>
                  
                  <Row form>
                    <Col md="6" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        幣種一
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormSelect onChange={evt => {setSelectedSymbol(evt.target.value)}}>
                            {
                              symbolArray1.map((s, i) => {
                                return (<option key={i}>{s}</option>);
                              })
                            }
                          </FormSelect>
                          {/*<FormInput value={selectedSymbol1Ratio} onChange={evt => {setSelectedSymbol1Ratio(evt.target.value)}}/>
                          <InputGroupAddon type="append">
                            <InputGroupText>%</InputGroupText>
                          </InputGroupAddon>*/}
                        </InputGroup>
                      </div>
                    </Col>
                    <Col md="6" className="form-group">
                      {/*<strong className="text-muted d-block mb-2">
                        幣種二
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormSelect onChange={evt => {setSelectedSymbol2(evt.target.value)}}>
                            {
                              symbolArray2.map((s, i) => {
                                return (<option key={i}>{s}</option>);
                              })
                            }
                          </FormSelect>
                          <FormInput value={selectedSymbol2Ratio} onChange={evt => {setSelectedSymbol2Ratio(evt.target.value)}}/>
                          <InputGroupAddon type="append">
                            <InputGroupText>%</InputGroupText>
                          </InputGroupAddon>
                        </InputGroup>
                          </div>*/}
                    </Col>
                  </Row>

                  <Row form>
                    <Col md="6" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        循環方式
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormSelect>
                            <option>單次循環</option>
                            <option>循環做單</option>
                          </FormSelect>
                        </InputGroup>
                      </div>
                    </Col>
                    <Col md="6" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        槓桿倍數
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormSelect>
                          <option>{selectedLeverage}</option>
                          </FormSelect>
                        </InputGroup>
                      </div>
                    </Col>
                  </Row>

                  <Row form>
                    <Col md="6" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        最大回撤
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormInput placeholder="0.1 ~ 1" value={maxDrawdown} onChange={evt => {setMaxDrawdown((parseFloat(evt.target.value, 10) > 1.0) ? '1' : evt.target.value )}}/>
                          <InputGroupAddon type="append">
                            <InputGroupText>{(maxDrawdown == '') ? '' : parseFloat(maxDrawdown, 10) * 100} %</InputGroupText>
                          </InputGroupAddon>
                        </InputGroup>
                      </div>
                    </Col>
                    <Col md="6" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        劃轉現貨
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormInput readOnly={(selectedExch === '幣安') ? (future ? false : true) : true}/>
                          <InputGroupAddon type="append">
                            <InputGroupText>USDT</InputGroupText>
                          </InputGroupAddon>
                        </InputGroup>
                      </div>
                    </Col>
                  </Row>

                  <Row form>
                  <Col className="form-group">
                      <strong className="text-muted d-block mb-2">
                        首單金額
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormInput readOnly={future ? true : false} value={maxAmount} onChange={(evt) => {setMaxAmount((parseFloat(evt.target.value, 10) > symbolBalance) ? symbolBalance : evt.target.value )}}/>
                          <InputGroupAddon type="append">
                            <InputGroupText>USDT</InputGroupText>
                          </InputGroupAddon>
                        </InputGroup>
                      </div>
                    </Col>
                    <Col className="form-group">
                      <strong className="text-muted d-block mb-2">
                        可用餘額
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormInput value={symbolBalance} readOnly/>
                          <InputGroupAddon type="append">
                            <InputGroupText>USDT</InputGroupText>
                          </InputGroupAddon>
                        </InputGroup>
                      </div>
                    </Col>
                  </Row>

                  <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                    <Button onClick={confirmNewTrading} outline theme="success" >
                      確定開通
                    </Button>
                  </div>
                </ListGroupItem>
              </ListGroup>
            </Card>
          </Col>
        </Row>
      )
       : null}

    {/* Default Light Table */}
    <Row>
      <Col>
        <Card small className="mb-4">
          <CardHeader className="border-bottom">
            <h6 className="m-0">已啟動機器人</h6>
          </CardHeader>
          <CardBody className="p-0 pb-3">
            <table className="table mb-0">
              <thead className="bg-light">
                <tr>
                  <th scope="col" className="border-0">
                    #
                  </th>
                  <th scope="col" className="border-0">
                    交易所
                  </th>
                  <th scope="col" className="border-0">
                    貨幣
                  </th>
                  <th scope="col" className="border-0">
                    均價
                  </th>
                  <th scope="col" className="border-0">
                    模式
                  </th>
                  <th scope="col" className="border-0">
                    首單金額
                  </th>
                </tr>
              </thead>
              <tbody>
                {
                  briefBotInfoArray.map((s, i) => {
                    return (
                    <tr key={i}>
                      <td>{s.brfIdx}</td>
                      <td>{s.brfExch}</td>
                      <td>{s.brfSymbol}</td>
                      <td>{s.brfAvgPrice}</td>
                      <td>{s.brfMode}</td>
                      <td>{s.brfQty}</td>
                    </tr>
                    );
                  })
                }

                {/*
                <tr>
                  <td>1</td>
                  <td>Ali</td>
                  <td>Kerry</td>
                  <td>Russian Federation</td>
                  <td>Gdańsk</td>
                  <td>107-0339</td>
                </tr>
                <tr>
                  <td>2</td>
                  <td>Clark</td>
                  <td>Angela</td>
                  <td>Estonia</td>
                  <td>Borghetto di Vara</td>
                  <td>1-660-850-1647</td>
                </tr>
                <tr>
                  <td>3</td>
                  <td>Jerry</td>
                  <td>Nathan</td>
                  <td>Cyprus</td>
                  <td>Braunau am Inn</td>
                  <td>214-4225</td>
                </tr>
                <tr>
                  <td>4</td>
                  <td>Colt</td>
                  <td>Angela</td>
                  <td>Liberia</td>
                  <td>Bad Hersfeld</td>
                  <td>1-848-473-7416</td>
                </tr>
                */}
              </tbody>
            </table>
          </CardBody>
        </Card>
      </Col>
    </Row>

  </Container>
);}

export default Tradings;
