import React from 'react';
import ReactDOM from 'react-dom';
import App from './App';

import * as serviceWorker from './serviceWorker';

import './index.css';
// Import the functions you need from the SDKs you need
import { initializeApp } from "firebase/app";
import { getAnalytics } from "firebase/analytics";
import { getFirestore } from "firebase/firestore";
// TODO: Add SDKs for Firebase products that you want to use
// https://firebase.google.com/docs/web/setup#available-libraries

// Your web app's Firebase configuration
// For Firebase JS SDK v7.20.0 and later, measurementId is optional
const firebaseConfig = {
  apiKey: "AIzaSyAy3q_rAZMBC3oJfR96KSbCVYMPdmu_IHs",
  authDomain: "trading-bot-d40d7.firebaseapp.com",
  projectId: "trading-bot-d40d7",
  storageBucket: "trading-bot-d40d7.appspot.com",
  messagingSenderId: "475314170291",
  appId: "1:475314170291:web:f2875a8ddf6902a85703c6",
  measurementId: "G-2J5QDBJ9VK"
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);
const analytics = getAnalytics(app);

export const firestoreDB = getFirestore(app);

ReactDOM.render(<App />, document.getElementById('root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: http://bit.ly/CRA-PWA
serviceWorker.unregister();

/*

*/

/*
<script type="module">
  // Import the functions you need from the SDKs you need
  import { initializeApp } from "https://www.gstatic.com/firebasejs/9.8.1/firebase-app.js";
  import { getAnalytics } from "https://www.gstatic.com/firebasejs/9.8.1/firebase-analytics.js";
  // TODO: Add SDKs for Firebase products that you want to use
  // https://firebase.google.com/docs/web/setup#available-libraries

  // Your web app's Firebase configuration
  // For Firebase JS SDK v7.20.0 and later, measurementId is optional
  const firebaseConfig = {
    apiKey: "AIzaSyAy3q_rAZMBC3oJfR96KSbCVYMPdmu_IHs",
    authDomain: "trading-bot-d40d7.firebaseapp.com",
    projectId: "trading-bot-d40d7",
    storageBucket: "trading-bot-d40d7.appspot.com",
    messagingSenderId: "475314170291",
    appId: "1:475314170291:web:f2875a8ddf6902a85703c6",
    measurementId: "G-2J5QDBJ9VK"
  };

  // Initialize Firebase
  const app = initializeApp(firebaseConfig);
  const analytics = getAnalytics(app);
</script>
*/
