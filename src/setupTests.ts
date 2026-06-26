import '@testing-library/jest-dom';

// react-router v7 uses TextEncoder which jsdom (Jest 27) doesn't provide
const { TextEncoder, TextDecoder } = require('util');
global.TextEncoder = TextEncoder;
global.TextDecoder = TextDecoder;
