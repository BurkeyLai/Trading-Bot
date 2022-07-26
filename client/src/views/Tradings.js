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
import {toast, ToastContainer} from 'react-toastify';
import Toast from "../utils/toastify";
import { RowComponent } from "../utils/rowComponent";
import { User, 
  Connect, 
  Message, 
  Exchange, 
  ExchangeConfig, 
  DepositAddresses, 
  MarketInfoRequest,
  AccountBalanceRequest,
  ClosePositionRequest,
  CreateOrderRequest,
  OrderInfoRequest, } from "../service_pb";
import { IsSignOut } from "../components/layout/MainNavbar/NavbarNav/UserActions";

const exchNum = 2, exch1 = "火幣", exch2 = "幣安", conNum = 2, type1 = "U本位", type2 = "幣本位";
const exchNameArray = ['huobi', 'binance'];
const modeNameArray = ['spot', 'future'];
const exchcfg = new ExchangeConfig();
const user = new User();
const connect = new Connect();
const msg = new Message();
var userRef = null;
var userSnap = null;
//             "homepage": "/static-tradingbot-client",
const Tradings = ({ client }) => {
    const { isSignOut } = React.useContext(IsSignOut);
    //const symbolArray1 = [], symbolArray2 = [];
    //const symbolTest = ['1', '2', '3'];
    const [briefBotInfoArray, setBriefBotInfoArray] = useState([]);
    const [detailBotInfoArray, setDetailBotInfoArray] = useState([]);
    //const [updateBriefBotInfoArray, setUpdateBriefBotInfoArray] = useState(false);

    const [symbolArray1, setSymbolArray1] = useState([]);
    const [symbolArray2, setSymbolArray2] = useState([]);
    const [selectedExch, setSelectedExch] = useState(exch1);
    const [selectedType, setSelectedType] = useState(type1);
    const [selectedSymbol, setSelectedSymbol] = useState('');
    //const [selectedSymbol2, setSelectedSymbol2] = useState('');
    const [selectedLeverage, setSelectedLeverage] = useState('1');
    const [maxDrawdown, setMaxDrawdown] = useState('');
    const [coverPosition, setCoverPosition] = useState('1');
    const [maxAmount, setMaxAmount] = useState(0);
    const [future, setFuture] = useState(false);
    const [newTrading, setNewTrading] = useState(false);

    const [selectedStrategy, setSelectedStrategy] = useState(0);
    const [symbolBalance, setSymbolBalance] = useState('');
    const [dropPercent, setDropPercent] = useState('3');
    const [goUpPercent, setGoUpPercent] = useState('5');
    const [expandTableDetails, setExpandTableDetails] = useState(0);
    const [lastTableIndex, setLastTableIndex] = useState(0);
    const [orderIdArray, setOrderIdArray] = useState([]);
    const [orderIdName, setOrderIdName] = useState('');
    const [referralData, setReferralData] = useState([]);
    const [cycleType, setCycleType] = useState('單次循環');
    const [isClosePosition, setIsClosePosition] = useState(false);
    const [confirmClosePosition, setConfirmClosePosition] = useState('');
    
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

    const askClosePosition = () => {
      if (user.getId() !== '') {
        const req = new ClosePositionRequest();
        const mode = detailBotInfoArray[0].dtlMode;
        const exch = detailBotInfoArray[0].dtlExchange;
        const symbol = detailBotInfoArray[0].dtlSymbol;
        msg.setContent("Ask for close position...");
        msg.setTimestamp(new Date().toLocaleTimeString());
        req.setMsg(msg);
        if (detailBotInfoArray[0] != null) {
          req.setMode(mode);
          (exch === 'huboi') ? req.setExchange('Huobi') : req.setExchange('Binance');
          req.setSymbol(symbol);
        }
        client.closePosition(req, null, (err, resp) => {
          if (resp) {
            console.log(resp.getContent())
            if (resp.getContent() != '') {
              //shutDownBot(symbol);
              Toast('success', true, "平倉成功");
            }
          }
        });
      }
    }

    const shutDownBot = (symbol) => {
      if (userRef != null) { 
        const storeBots = async () => {
          userSnap = await getDoc(userRef);
          if (userSnap.exists) {
            if (userSnap.get("bots_array")) {
              var arr = userSnap.get("bots_array");
              for (var i = 0; i < arr.length; ++i) {
                if(arr[i].symbol == symbol) {
                  arr.splice(i, 1);
                }
              }
              setDoc(userRef, {
                'bots_array': arr
              }, { merge: true });
            }
          }
          return '平倉成功'
        }

        storeBots().then((res) => {
          console.log(res);
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
        req.setDroppercent((parseFloat(dropPercent, 10) / 100).toString());
        req.setGouppercent((parseFloat(goUpPercent, 10) / 100).toString());
        (selectedExch === '火幣') ? req.setExchange('Huobi') : req.setExchange('Binance');
        (selectedType === 'U本位') ? req.setType('usdt') : req.setType('coin');
        req.setSymbol(selectedSymbol);
        (cycleType === '單次循環') ? req.setCycletype('single') : req.setCycletype('recursive');
        req.setLeverage(selectedLeverage);
        req.setMaxdrawdown(maxDrawdown);
        req.setWithdrawspot('');
        req.setQuantity(maxAmount.toString());
        req.setCoverposition(coverPosition);
        client.createOrder(req, null, (err, resp) => {
          if (resp != null) {
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
                      m_type : req.getType(),
                      symbol : strs[0] + strs[1],
                      symbol_balance : '',
                      cycle_type : req.getCycletype(),
                      leverage : req.getLeverage(),
                      max_drawdown : req.getMaxdrawdown(),
                      withdraw_spot : req.getWithdrawspot(),
                      quantity : req.getQuantity(),
                      average_price: '',
                      order_id_list: [],
                      cover_position: req.getCoverposition()
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
                  return '儲存成功'
                }

                storeBots().then((res) => {
                  console.log(res);
                });
              }
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
      setSelectedStrategy(0);
      setDropPercent('3');
      setGoUpPercent('5');
    };
    const strategy2 = () => {
      setSelectedStrategy(1);
      setDropPercent('1');
      setGoUpPercent('1');
    };
    const strategy3 = () => {
      setSelectedStrategy(2);
      setDropPercent('');
      setGoUpPercent('');
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
      } else if (dropPercent === '') {
        Toast('error', true, '請輸入偵測％');
      } else if (goUpPercent === '') {
        Toast('error', true, '請輸入回彈％');
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
                  brfBalance: arr[i].symbol_balance,
                  brfMode: (arr[i].mode == 'spot') ? '現貨' : '合約',
                  //brfQty: arr[i].quantity
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
                  arr[i].symbol_balance = info.getSymbolbalance();
                  arr[i].quantity = info.getQuantity();
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
            //if (response.getBotinfo()) {
            //  updateBotInfo(response.getBotinfo());
            //}
            if (messageContent == 'First Bot Info!' || messageContent == 'Update Bot Info!') {
              updateBotInfo(response.getBotinfo());
              Toast('success', true, "下單成功");
            } else if (messageContent == 'Shut Down Bot!') {
              shutDownBot(response.getBotinfo().getModelname());
              Toast('error', true, "下單失敗，請重新下單");
            }
          });
        } else {
          // User is signed out
        }
      });
    }

    useEffect(() => {
      let isMounted = true;
      if(isMounted){
        askCreateStream();
        updateBriefBotInfo();
      }
      return () => {
        isMounted = false;
      };
    }, [])

    useEffect(() => {
      let isMounted = true;
      if(isMounted){
        askMarketSymbols();
        askAccountBalance();
      }
      return () => {
        isMounted = false;
        if (isSignOut) {
          setSelectedExch(exch1);
        }
      };
    }, [selectedExch])

    useEffect(() => {
      let isMounted = true;
      if(isMounted){
        //askExchKeyConfig();
        //setNewTrading(false);
        askMarketSymbols();
        askAccountBalance();
        if (future) {
          setSelectedLeverage('20');
        } else {
          setSelectedLeverage('1');
        }
      }
      return () => {
        isMounted = false;
        if (isSignOut) {
          setFuture(false);
        }
      };
    }, [future])

    useEffect(() => {
      let isMounted = true;
      if (isMounted && user.getId() !== '' && orderIdName != '請選擇') {
        const req = new OrderInfoRequest();
        msg.setContent("Ask for order information...");
        msg.setTimestamp(new Date().toLocaleTimeString());
        req.setMsg(msg);
        //(future) ? req.setMode('future') : req.setMode('spot');
        if (detailBotInfoArray[0] != null) {
          req.setMode(detailBotInfoArray[0].dtlMode);
          (detailBotInfoArray[0].dtlExchange === 'huboi') ? req.setExchange('Huobi') : req.setExchange('Binance');
          req.setSymbol(detailBotInfoArray[0].dtlSymbol);
        }
        req.setOrderid(orderIdName);
        client.orderInfo(req, null, (err, resp) => {
          if (resp) {
            //console.log(resp.getTimestamp())
            //var t = new Date( resp.getTimestamp()/1000 );
            //var formatted = moment(t).format("dd.mm.yyyy hh:MM:ss");
            //console.log(formatted)
            const qty = {
              title: '成交成本',
              value: resp.getQuantity()
            }
            const amount = {
              title: '成交量',
              value: resp.getAmount()
            }
            const price = {
              title: '成交價格',
              value: resp.getPrice()
            }
            const type = {
              title: '成交方式',
              value: (resp.getType() == 'MARKET') ? '市價' : '限價'
            }
            const side = {
              title: '買/賣',
              value: (resp.getSide() == 'BUY') ? '買' : '賣'
            }
            const time = {
              title: '成交時間',
              value: resp.getTimestamp()
            }
            setReferralData([qty, amount, price, type, side, time]);
          }
        });
      }
      return () => {
        isMounted = false;
        if (isSignOut) {
          setOrderIdName('');
        }
      };
    }, [orderIdName])

    const fetchDetails = (data) => {
      setOrderIdName('');
      if (expandTableDetails != 0 && data.brfIdx == lastTableIndex) {
        setExpandTableDetails(0);
      } else {
        setExpandTableDetails(data.brfIdx + 1);
      }
      setLastTableIndex(data.brfIdx);

      const auth = getAuth();
      onAuthStateChanged(auth, async (f_user) => {
        if (f_user) {
          userSnap = await getDoc(userRef);
          if (userSnap.exists) {
            if (userSnap.get("bots_array")) {
              var arr = userSnap.get("bots_array");
              var updateArray = [];
              for (var i = 0; i < arr.length; ++i) {
                if (i == data.brfIdx /*&& arr[i].order_id_list != null*/) {
                  const l = (arr[i].order_id_list != null) ? arr[i].order_id_list.length : 0;
                  const tbl = 
                  {
                    dtlIdx: i,
                    dtlActive: (arr[i].bot_active) ? '是' : '否',
                    dtlOrderNum: (l == 0) ? ('無') : ((l == 1) ? '首倉' : '補 ' + (l - 1).toString()),
                    dtlLeverage: arr[i].leverage,
                    dtlCycle: (arr[i].cycle_type == 'single') ? '單次循環' : '循環做單',
                    dtlMaxDrawdown: arr[i].max_drawdown,
                    dtlCoverPosition: arr[i].cover_position,
                    dtlQty: arr[i].quantity,
                    dtlDrop: (parseFloat(arr[i].drop_percent, 10) * 100).toString(),
                    dtlGoUp: (parseFloat(arr[i].go_up_percent, 10) * 100).toString(),
                    dtlExchange: arr[i].exchange,
                    dtlSymbol: arr[i].symbol,
                    dtlMode: arr[i].mode
                  };
                  updateArray = [...updateArray, tbl];

                  if (arr[i].order_id_list != null) {
                    setOrderIdArray(['請選擇', ...arr[i].order_id_list]);
                  } else {
                    setOrderIdArray(['請選擇']);
                  }
                  break;
                }
              }
              setDetailBotInfoArray(updateArray);
            }
          }
        }
      });
    }
  
    const renderResultRows = (data) => {
      //console.log(data)
      return (
        <RowComponent
          key={data.brfIdx}
          data={data}
          onClick={fetchDetails}
        />
      )
    }
    /*
    const askOrderInfo = (id) => {
      if (user.getId() !== '') {
        const req = new OrderInfoRequest();
        msg.setContent("Ask for order information...");
        msg.setTimestamp(new Date().toLocaleTimeString());
        req.setMsg(msg);
        (future) ? req.setMode('future') : req.setMode('spot');
        (detailBotInfoArray[0].dtlExchange === 'huboi') ? req.setExchange('Huobi') : req.setExchange('Binance');
        req.setSymbol(detailBotInfoArray[0].dtlSymbol);
        req.setOrderid(id);
        client.orderInfo(req, null, (err, resp) => {
          if (resp) {
            const qty = {
              title: '成交成本',
              value: resp.getQuantity()
            }
            const amount = {
              title: '成交量',
              value: resp.getAmount()
            }
            const price = {
              title: '成交價格',
              value: resp.getPrice()
            }
            const type = {
              title: '成交方式',
              value: (resp.getType() == 'MARKET') ? '市價' : '限價'
            }
            const side = {
              title: '買/賣',
              value: (resp.getSide() == 'BUY') ? '買' : '賣'
            }
            setReferralData([qty, amount, price, type, side]);
          }
        });
      }

      setOrderIdName(id);
    }
    */
    
    
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
        <ButtonGroup className="mb-3 mr3">
          <Button onClick={() => {
            askCreateStream();
            updateBriefBotInfo();
          }} outline theme="success">
            重新整理
          </Button>
        </ButtonGroup>
      </Col>
    </Row>
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
                          <Button onClick={strategy1} outline={selectedStrategy == 0 ? false : true} theme="success">
                            保守
                          </Button>
                          <Button onClick={strategy2} outline={selectedStrategy == 1 ? false : true} theme="success">
                            激進
                          </Button>
                          <Button onClick={strategy3} outline={selectedStrategy == 2 ? false : true} theme="success">
                            自訂
                          </Button>
                        </ButtonGroup>
                      </div>
                    </Col>
                  </Row>
                  
                  <Row form>
                    <Col md="6" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        偵測％
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormInput readOnly={selectedStrategy == 2 ? false : true} placeholder="0.1 ~ 100" value={dropPercent} onChange={evt => {setDropPercent((parseFloat(evt.target.value, 10) > 100) ? '100' : (parseFloat(evt.target.value, 10) < 0.1 ? '0.1' : evt.target.value))}}/>
                          <InputGroupAddon type="append">
                            <InputGroupText>%</InputGroupText>
                          </InputGroupAddon>
                        </InputGroup>
                      </div>
                    </Col>
                    <Col md="6" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        回彈％
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormInput readOnly={selectedStrategy == 2 ? false : true} placeholder="0.1 ~ 100" value={goUpPercent} onChange={evt => {setGoUpPercent((parseFloat(evt.target.value, 10) > 100) ? '100' : (parseFloat(evt.target.value, 10) < 0.1 ? '0.1' : evt.target.value))}}/>
                          <InputGroupAddon type="append">
                            <InputGroupText>%</InputGroupText>
                          </InputGroupAddon>
                        </InputGroup>
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
                        幣種
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
                      <strong className="text-muted d-block mb-2">
                        補倉策略
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormInput placeholder="1 ~ 2" value={coverPosition} onChange={evt => {setCoverPosition((parseFloat(evt.target.value, 10) > 2.0) ? '2' : ((parseFloat(evt.target.value, 10) < 1.0) ? '1' : evt.target.value))}}/>
                          <InputGroupAddon type="append">
                            <InputGroupText>{(coverPosition == '') ? '' : parseFloat(coverPosition, 10) * 100} %</InputGroupText>
                          </InputGroupAddon>
                        </InputGroup>
                      </div>
                    </Col>
                  </Row>

                  <Row form>
                    <Col md="6" className="form-group">
                      <strong className="text-muted d-block mb-2">
                        循環方式
                      </strong>
                      <div>
                        <InputGroup className="mb-2">
                          <FormSelect onChange={evt => {
                              setCycleType(evt.target.value);
                            }}>
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
                    持有數量
                  </th>
                  <th scope="col" className="border-0">
                    模式
                  </th>
                </tr>
              </thead>
              <tbody>
                {
                  
                  briefBotInfoArray.map((s, i) => {
                    return renderResultRows(s)
                  })
                  
                }
              </tbody>
            </table>
          </CardBody>
        </Card>
        {expandTableDetails != 0 ?
        (<Card small className="mb-4 overflow-hidden">
          <CardHeader className="bg-dark">
            <h6 className="m-0 text-white">詳細資料</h6>
          </CardHeader>
          <CardBody className="p-0 pb-3 bg-dark">
            <table className="table table-dark mb-0">
              <thead className="thead-dark">
                <tr>
                  <th scope="col" className="border-0">
                    #
                  </th>
                  <th scope="col" className="border-0">
                    已啟動
                  </th>
                  <th scope="col" className="border-0">
                    購買單數
                  </th>
                  <th scope="col" className="border-0">
                    槓桿倍數
                  </th>
                  <th scope="col" className="border-0">
                    循環方式
                  </th>
                  <th scope="col" className="border-0">
                    最大回撤
                  </th>
                  <th scope="col" className="border-0">
                    補倉策略
                  </th>
                  <th scope="col" className="border-0">
                    下單金額
                  </th>
                  <th scope="col" className="border-0">
                    偵測％
                  </th>
                  <th scope="col" className="border-0">
                    回彈％
                  </th>
                </tr>
              </thead>
              <tbody>
                {
                  
                  detailBotInfoArray.map((s, i) => {
                    return (
                      <tr key={i}>
                        <td>{s.dtlIdx}</td>
                        <td>{s.dtlActive}</td>
                        <td>{s.dtlOrderNum}</td>
                        <td>{s.dtlLeverage}</td>
                        <td>{s.dtlCycle}</td>
                        <td>{s.dtlMaxDrawdown}</td>
                        <td>{s.dtlCoverPosition}</td>
                        <td>{s.dtlQty}</td>
                        <td>{s.dtlDrop}</td>
                        <td>{s.dtlGoUp}</td>
                      </tr>
                    )
                  })
                  
                }
              </tbody>
            </table>
          </CardBody>
          <CardHeader className="bg-dark">
            <h6 className="m-0 text-white">訂單</h6>
          </CardHeader>
          <CardBody className="p-0 pb-3 bg-dark">
            <ListGroup flush>
              <ListGroupItem className="px-3 bg-dark">
                <Row form>
                  <Col md="6" className="form-group">
                    <strong className="text-muted d-block mb-2">
                      單號
                    </strong>
                    <div>
                      <InputGroup className="mb-2">
                        <FormSelect onChange={evt => {setOrderIdName(evt.target.value)}}>
                          {
                            orderIdArray.map((s, i) => {
                              return (<option key={i}>{s}</option>);
                            })
                          }
                        </FormSelect>
                      </InputGroup>
                    </div>
                  </Col>
                  <Col md="6" className="form-group">
                    <strong className="text-muted d-block mb-2">
                      個單數據
                    </strong>
                    <div>
                      <CardBody className="p-0">
                        <ListGroup small flush className="list-group-small">
                          {orderIdName != '' && orderIdName != '請選擇' ? 
                            referralData.map((item, idx) => (
                            <ListGroupItem key={idx} className="d-flex px-3">
                              <span className="text-semibold text-fiord-blue">{item.title}</span>
                              <span className="ml-auto text-right text-semibold text-reagent-gray">
                                {item.value}
                              </span>
                            </ListGroupItem>
                          )) : null}
                        </ListGroup>
                      </CardBody>
                    </div>
                  </Col>
                </Row>
                
                <Row form>
                  <Col md="6" className="form-group">
                    <div style={{ display: 'flex', justifyContent: 'flex-start' }}>
                      <Button onClick={() => {
                        isClosePosition ? setIsClosePosition(false) : setIsClosePosition(true)
                        }} outline theme="success" >
                        一鍵平倉
                      </Button>
                      {isClosePosition ?
                      <InputGroup className="ml-2">
                        <FormInput placeholder="請輸入'YES'" onChange={evt => {setConfirmClosePosition(evt.target.value)}}/>
                        <InputGroupAddon type="append">
                          <Button onClick={() => {
                            if (confirmClosePosition == 'YES') {
                              //console.log('Close Position')
                              askClosePosition();
                              setIsClosePosition(false);
                            }
                          }} outline theme="success">
                            確定平倉
                          </Button>
                        </InputGroupAddon>
                      </InputGroup>
                      : null
                      }
                    </div>
                  </Col>
                </Row>
                
                
              </ListGroupItem>
            </ListGroup>
          </CardBody>
        </Card>)
        : null}
      </Col>
    </Row>

    </Container>
);}

export default Tradings;
