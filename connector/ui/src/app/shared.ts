const kongAddress = process.env.NEXT_PUBLIC_KONG_ADDRESS??"";
const connectorKeyID = process.env.NEXT_PUBLIC_CONNECTOR_KEY_ID??"";
let password = process.env.NEXT_PUBLIC_PASSWORD ?? '';
let loadedFromLocalStorage = false;

const getPassword = () => {
  if (
    password === '' &&
    !loadedFromLocalStorage &&
    typeof localStorage !== 'undefined'
  ) {
    loadedFromLocalStorage = true;
    const storedPassword = localStorage.getItem('password');
    if (storedPassword) {
      password = storedPassword;
    }
  }
  return password;
};

const setPassword = (newPassword: string) => {
  password = newPassword;
  if (typeof localStorage !== 'undefined' && password !== '') {
    localStorage.setItem('password', password);
  }
};

const shared = {
  kongAddress,
  connectorKeyID,
  getPassword,
  setPassword,
};

export default shared;
