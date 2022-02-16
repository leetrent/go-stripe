select 
	o.id as OrderID,
	c.id as CustomerID,
	c.last_name,
	c.first_name
from 
	orders o
inner join
	customers c on c.id = o.customer_id 
where 
	o.widget_id = 2;