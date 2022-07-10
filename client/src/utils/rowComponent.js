import React from 'react';

export function RowComponent(props) {
  const { data, onClick } = props;

  function handleClick() {
    onClick(data);
  }

  return (
    <tr onClick={handleClick}>
      <td data-title="Index">{data.brfIdx}</td>
      <td data-title="Exchange">{data.brfExch}</td>
      <td data-title="Symbol">{data.brfSymbol}</td>
      <td data-title="Price">{data.brfAvgPrice}</td>
      <td data-title="Balance">{data.brfBalance}</td>
      <td data-title="Mode">{data.brfMode}</td>
    </tr>
  );
}